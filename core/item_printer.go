package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/juju/ansiterm"
)

type PrintFormat int

const (
	HumanPrintFormat PrintFormat = 0
	TextPrintFormat  PrintFormat = 1
	JSONPrintFormat  PrintFormat = 2

	ColMinWidth    int = 2  // minimum column width
	ColSpacesLen   int = 3  // how many spaces between columns (inc newline col)
	LargestDateLen int = 12 // length of "- Wed Sep 23"

	PrintTagLimit int = 4
)

func GetPrintFormatFromContext(ctx context.Context) PrintFormat {
	format := ctx.Value("format")
	if format == nil {
		format = "text"
	}
	return GetPrintFormat(format.(string))
}

func GetPrintFormat(format string) PrintFormat {
	return map[string]PrintFormat{
		"human": HumanPrintFormat,
		"text":  TextPrintFormat,
		"json":  JSONPrintFormat,
	}[format]
}

type ItemPrinter struct {
	bumpedColor  color.Attribute
	failColor    color.Attribute
	waitColor    color.Attribute
	successColor color.Attribute
	noteColor    color.Attribute

	hasher     hash.Hash32
	colorWheel map[int]int

	PrintFormat PrintFormat
}

func NewItemPrinter(ctx context.Context) *ItemPrinter {
	base := &ItemPrinter{
		bumpedColor:  color.FgYellow,
		failColor:    color.FgRed,
		successColor: color.FgGreen,
		waitColor:    color.FgWhite,
		noteColor:    color.FgBlue,
		PrintFormat:  GetPrintFormatFromContext(ctx),
	}

	base.hasher = fnv.New32a()

	// there are 216 non "standard" colors
	// some of them might be hard to read on a regular terminal, so we limit
	// this is pretty arbitrary based on the scheme I'm currently usng
	colorWheel := map[int]int{}
	for i := 33; i < 52; i++ {
		colorWheel[len(colorWheel)] = i
	}
	for i := 69; i < 88; i++ {
		colorWheel[len(colorWheel)] = i
	}
	for i := 99; i < 231; i++ {
		colorWheel[len(colorWheel)] = i
	}

	base.colorWheel = colorWheel
	return base
}

func (ip *ItemPrinter) Print(items ...*Item) {
	ip.FPrint(os.Stdout, items...)
}

func (ip *ItemPrinter) PrintSingleWithConnected(item *Item, connections ...*Item) {
	switch ip.PrintFormat {
	case HumanPrintFormat:
		baseColor := color.New(color.Bold)
		baseColor.EnableColor()
		fmt.Fprint(os.Stdout, baseColor.Sprintf("Main Item:\n\n"))
		ip.FPrint(os.Stdout, item)
		fmt.Fprint(os.Stdout, baseColor.Sprintln("Connected Items:"))
		ip.FPrint(os.Stdout, connections...)
	default:
		ip.FPrint(os.Stdout, item)
		ip.FPrint(os.Stdout, connections...)
	}
}

func (ip *ItemPrinter) FPrint(w io.Writer, items ...*Item) {
	if len(items) == 0 {
		return
	}

	tw := ansiterm.NewTabWriter(w, ColMinWidth, 0, 1, ' ', 0)
	defer tw.Flush()

	if len(items) == 1 {
		switch ip.PrintFormat {
		case TextPrintFormat:
			ip.fPrintItemCompact(w, items[0])
		case HumanPrintFormat:
			ip.fPrintItemDetail(tw, items[0])
		case JSONPrintFormat:
			ip.fPrintItemJSON(tw, items[0])
		}
		return
	}

	currDay := items[0].Time().Day() - 1 // set to something different
	currYear := items[0].Time().Year()
	showYear := false
	maxTagStringLength := ip.maxTagStringLength(items)

	for _, item := range items {
		if item.Time().Year() != currYear {
			showYear = true
			currYear = 1990 // reset so we can always print first header
			break
		}
	}

	for _, item := range items {
		switch ip.PrintFormat {
		case TextPrintFormat:
			ip.fPrintItemCompact(w, item)
		case HumanPrintFormat:
			// new year so print header if we should
			if showYear && (currYear != item.Time().Year()) {
				fmt.Fprintf(tw, "\t\t\n")
				fmt.Fprintf(tw, "### %s\t\t\n", item.Time().Format("2006"))
				currYear = item.Time().Year()
			}
			// new day so print header
			if currDay != item.Time().Day() {
				fmt.Fprintf(tw, "\t\t\n")
				fmt.Fprintf(tw, "- %s\t\t\n", item.Time().Format("Mon Jan 02"))
				currDay = item.Time().Day()
			}
			ip.fPrintItemHuman(tw, item, maxTagStringLength)
		case JSONPrintFormat:
			ip.fPrintItemJSON(tw, item)
		}
	}
}

func (ip *ItemPrinter) fPrintItemDetail(w io.Writer, item *Item) {
	fmt.Fprintf(w, "%s -- %v\n", ip.doneStatus(item), item.Time().Format("Mon, 02 Jan 2006 15:04:05"))
	baseColor := color.New(color.Bold)
	baseColor.EnableColor()
	fmt.Fprintf(w, "Kind: %v\n", item.Kind())
	if item.NextID() != "" {
		fmt.Fprintf(w, "Bumped to: %s\n", baseColor.Sprintf("%s", item.NextID()))
	}
	if item.PreviousID() != "" {
		fmt.Fprintf(w, "Bumped from: %s\n", baseColor.Sprintf("%s", item.PreviousID()))
	}
	if len(item.Tags()) != 0 {
		fmt.Fprintf(w, "Tags: %v\n", baseColor.Sprintf("%s", ip.itemTags(item, true)))
	}
	if len(item.Connections()) != 0 {
		fmt.Fprintf(w, "Connections: %v\n", baseColor.Sprintf("%s", item.Connections()))
	}
	fmt.Fprintf(w, "Data:\n%s\n\n", item.Data())
}

func (ip *ItemPrinter) fPrintItemCompact(w io.Writer, item *Item) {
	refID := ""
	if item.NextID() != "" {
		refID = "->" + item.NextID()
	}
	if item.PreviousID() != "" {
		refID = "<-" + item.PreviousID()
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%v\n", item.ID(), item.internalID, item.Status(), refID, item.Data(), item.Time().Format(time.RFC3339), item.Kind())
}

type JSONItem struct {
	ID         string
	InternalID string
	NextID     string
	PreviousID string

	Data       string
	Status     string
	TimeString string
	Tags       []string
	Kind       string
}

func (ip *ItemPrinter) fPrintItemJSON(w io.Writer, item *Item) {
	tags := item.Tags()
	tagStrings := []string{}
	for _, tag := range tags {
		tagStrings = append(tagStrings, tag.Name())
	}

	jsonItem := JSONItem{
		ID:         item.ID(),
		InternalID: item.internalID,
		NextID:     item.NextID(),
		PreviousID: item.PreviousID(),
		Data:       item.Data(),
		Status:     item.Status(),
		TimeString: item.Time().Format(time.RFC3339),
		Tags:       tagStrings,
		Kind:       item.Kind().String(),
	}
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(jsonItem)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "%s", buf.String())
}

func (ip *ItemPrinter) fPrintItemHuman(w io.Writer, item *Item, maxTagStringLength int) {
	dataString := TrimString(strings.Split(item.Data(), "\n")[0], LargestDateLen+ColSpacesLen+ColMinWidth+maxTagStringLength)
	fmt.Fprintf(w, "%s\t%s\t%s\t\n", ip.doneStatus(item), dataString, ip.itemTags(item, false))
}

func (ip *ItemPrinter) itemTags(item *Item, showAll bool) string {
	if len(item.Tags()) == 0 {
		return ""
	}

	tagStrings := []string{}
	for i, tag := range item.Tags() {
		if i > PrintTagLimit && !showAll {
			tagStrings = append(tagStrings, "\u2026")
			break
		}
		ip.hasher.Write([]byte(tag.Name()))
		num := int(ip.hasher.Sum32())

		// find the appropriate color in our wheel
		num = num % len(ip.colorWheel)

		//terminal escape codes are in the format: 38;5;n for the larger range of colors
		tagStrings = append(tagStrings, ip.tagColor(tag.Name(), []int{38, 5, ip.colorWheel[num]}))
		ip.hasher.Reset()
	}

	return fmt.Sprintf("%s", tagStrings)
}

func (ip *ItemPrinter) maxTagStringLength(items []*Item) int {
	maxLength := 0
	for _, item := range items {
		if len(ip.rawItemTagsString(item)) > maxLength {
			maxLength = len(ip.rawItemTagsString(item))
		}
	}
	return maxLength
}

func (ip *ItemPrinter) rawItemTagsString(item *Item) string {
	if len(item.Tags()) == 0 {
		return ""
	}

	tagStrings := []string{}
	for i, tag := range item.Tags() {
		if i > PrintTagLimit {
			tagStrings = append(tagStrings, "\u2026")
			break
		}
		tagStrings = append(tagStrings, tag.Name())
	}

	return fmt.Sprintf("%s", tagStrings)
}

func (ip *ItemPrinter) tagColor(tagName string, params []int) string {
	format := make([]string, len(params))
	for i, v := range params {
		format[i] = strconv.Itoa(int(v))
	}

	sequence := strings.Join(format, ";")
	return fmt.Sprintf("%s[%sm%s%s[%dm", "\x1b", sequence, tagName, "\x1b", 0)
}

func (ip *ItemPrinter) doneStatus(item *Item) string {
	switch item.Status() {
	case NoStatus:
		baseColor := color.New(ip.noteColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("⇒ %v", item.ID())
	case BumpedStatus:
		baseColor := color.New(ip.bumpedColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("⇒ %v", item.ID())
	case DoneStatus:
		baseColor := color.New(ip.successColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("✔ %v", item.ID())
	case WaitingStatus:
		if item.PreviousID() != "" { // i.e. was bumped
			baseColor := color.New(ip.bumpedColor)
			baseColor.EnableColor()
			return fmt.Sprintf("%s %v", baseColor.Sprintf("⇒"), item.ID()) // just color the arrow
		}
		baseColor := color.New(ip.waitColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("⇒ %v", item.ID())
	case SkippedStatus:
		baseColor := color.New(ip.failColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("✘ %v", item.ID())
	default:
		baseColor := color.New(ip.waitColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("? %v", item.ID())
	}
}
