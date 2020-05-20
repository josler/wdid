package core

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/josler/wdid/parser"
)

type ItemCreator struct {
	ctx context.Context
}

func (ic *ItemCreator) Create(data string, at time.Time) (*Item, error) {
	store := ic.ctx.Value("store").(Store)
	item := NewTask(data, at)
	err := ic.GenerateAndSaveMetadata(item)
	if err != nil {
		return nil, err
	}
	err = store.Save(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (ic *ItemCreator) maybeNewTime(item *Item, timeString string) error {
	if timeString != "" {
		span, err := TimeParser{Input: timeString}.Parse()
		if err != nil {
			return err
		}
		item.datetime = span.Start
	}
	return nil
}

func (ic *ItemCreator) maybeNewDescription(item *Item, data string) error {
	if data != "" {
		item.data = data
	}
	return nil
}

func (ic *ItemCreator) Edit(item *Item, data string, timeString string) (*Item, error) {
	store := ic.ctx.Value("store").(Store)
	err := ic.maybeNewTime(item, timeString)
	if err != nil {
		return nil, err
	}
	err = ic.maybeNewDescription(item, data)
	if err != nil {
		return nil, err
	}

	err = ic.GenerateAndSaveMetadata(item)
	if err != nil {
		return nil, err
	}
	err = store.Save(item)
	return item, err
}

func (ic *ItemCreator) GenerateAndSaveMetadata(item *Item) error {
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(item.Data())
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to generate tags, %v\n", err)
		}
		return err
	}
	err = ic.saveNewTags(item, tokenResult)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to save tags, %v\n", err)
		}
		return err
	}
	item.tags = []*Tag{}
	for _, resultTag := range tokenResult.Tags {
		item.tags = append(item.tags, NewTag(resultTag))
	}
	if len(tokenResult.Connections) > 0 {
		item.connections = tokenResult.Connections
	}
	return nil
}

func (ic *ItemCreator) Delete(item *Item) error {
	store := ic.ctx.Value("store").(Store)
	err := store.Delete(item)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to delete item. %v\n", err)
		}
		return err
	}
	return nil
}

func (ic *ItemCreator) saveNewTags(item *Item, tokenResult *parser.TokenResult) error {
	store := ic.ctx.Value("store").(Store)
	for _, resultTag := range tokenResult.Tags {
		tag := NewTag(resultTag)
		err := store.SaveTag(tag)
		if err != nil {
			return err
		}
		if ic.isVerbose() {
			fmt.Printf("saved tag %s\n", tag.Name())
		}
	}
	return nil
}

func (ic *ItemCreator) isVerbose() bool {
	v := ic.ctx.Value("verbose")
	return v != nil && v.(bool)
}

// GenerateID generates a 6 digit ID - with the last three digits sortable by Year, Month, Day
// Format: RRRYMD - where R is a random base36 rune.
// This means, that in any given day, there's 36^3 chance of a random collision - acceptable for this.
func GenerateID(at time.Time) string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return fmt.Sprintf("%c%c%c%s", Base36(r1.Intn(35)), Base36(r1.Intn(35)), Base36(r1.Intn(35)), IDSuffixForDate(at))
}

// numbers above 35 translate to the value of remainder. i.e. 36 is 0, 71 is z, 72 is 0...
func Base36(in int) rune {
	charMap := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	return charMap[(in % 36)]
}

func IDSuffixForDate(t time.Time) string {
	return fmt.Sprintf("%c%c%c", Base36(int(t.Year()-2000)), Base36(int(t.Month())), Base36(int(t.Day())))
}
