package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

type PrintFormat int

const (
	HumanPrintFormat PrintFormat = 0
	TextPrintFormat  PrintFormat = 1
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
	}[format]
}

type ItemPrinter struct {
	bumpedColor  color.Attribute
	failColor    color.Attribute
	waitColor    color.Attribute
	successColor color.Attribute

	PrintFormat PrintFormat
}

func NewItemPrinter(ctx context.Context) *ItemPrinter {
	return &ItemPrinter{bumpedColor: color.FgYellow, failColor: color.FgRed, successColor: color.FgGreen, waitColor: color.FgWhite, PrintFormat: GetPrintFormatFromContext(ctx)}
}

func (ip *ItemPrinter) Print(items ...*Item) {
	ip.FPrint(os.Stdout, items...)
}

func (ip *ItemPrinter) FPrint(w io.Writer, items ...*Item) {
	if len(items) == 0 {
		return
	}

	tw := &tabwriter.Writer{}
	tw.Init(w, 7, 0, 1, ' ', 0)
	defer tw.Flush()

	if len(items) == 1 {
		switch ip.PrintFormat {
		case TextPrintFormat:
			ip.fPrintItemCompact(w, items[0])
		case HumanPrintFormat:
			ip.fPrintItemDetail(tw, items[0])
		}
		return
	}

	currDay := items[0].Time().Day() - 1 // set to something different
	for _, item := range items {
		switch ip.PrintFormat {
		case TextPrintFormat:
			ip.fPrintItemCompact(w, item)
		case HumanPrintFormat:
			if currDay != item.Time().Day() {
				fmt.Fprintf(tw, "\t\t\t\n")
				fmt.Fprintf(tw, "- %s\t\t\t\n", item.Time().Format("Monday January 02"))
				currDay = item.Time().Day()
			}
			ip.fPrintItem(tw, item)
		}
	}
}

func (ip *ItemPrinter) fPrintItemDetail(w io.Writer, item *Item) {
	fmt.Fprintf(w, "%s -- %v\n", ip.doneStatus(item), item.Time().Format("Mon, 02 Jan 2006 15:04:05"))
	fmt.Fprintf(w, "InternalID: %s\n", item.internalID)
	baseColor := color.New(color.Bold)
	baseColor.EnableColor()
	if item.NextID() != "" {
		fmt.Fprintf(w, "Bumped to: %s\n", baseColor.Sprintf("%s", item.NextID()))
	}
	if item.PreviousID() != "" {
		fmt.Fprintf(w, "Bumped from: %s\n", baseColor.Sprintf("%s", item.PreviousID()))
	}
	if len(item.Tags()) != 0 {
		fmt.Fprintf(w, "Tags: %v\n", baseColor.Sprintf("%s", item.Tags()))
	}
	fmt.Fprintf(w, "Data:\n %s\n\n", item.Data())
}

func (ip *ItemPrinter) fPrintItemCompact(w io.Writer, item *Item) {
	refID := ""
	if item.NextID() != "" {
		refID = "->" + item.NextID()
	}
	if item.PreviousID() != "" {
		refID = "<-" + item.PreviousID()
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%q\t%s\n", item.ID(), item.internalID, item.Status(), refID, item.Data(), item.Time().Format(time.RFC3339))
}

func (ip *ItemPrinter) fPrintItem(w io.Writer, item *Item) {
	fmt.Fprintf(w, "%s\t%q\t%s\t%v\t\n", ip.doneStatus(item), TrimString(item.Data(), 20), ip.itemTags(item), item.Time().Format("15:04"))
}

func (ip *ItemPrinter) itemTags(item *Item) string {
	if len(item.Tags()) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", item.Tags())
}

func (ip *ItemPrinter) doneStatus(item *Item) string {
	switch item.Status() {
	case BumpedStatus:
		baseColor := color.New(ip.bumpedColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("⇒ %v", item.ID())
	case DoneStatus:
		baseColor := color.New(ip.successColor)
		baseColor.EnableColor()
		return baseColor.Sprintf("✔ %v", item.ID())
	case WaitingStatus:
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
