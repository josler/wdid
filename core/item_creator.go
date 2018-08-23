package core

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gitlab.com/josler/wdid/parser"
)

type ItemCreator struct {
	ctx context.Context
}

func (ic *ItemCreator) Create(data string, at time.Time) (*Item, error) {
	store := ic.ctx.Value("store").(Store)
	item := NewItem(data, at)
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(data)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to generate tags, %v\n", err)
		}
		return nil, err
	}
	err = ic.saveNewTags(item, tokenResult)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to save tags, %v\n", err)
		}
		return nil, err
	}
	item.tags = []*Tag{}
	for _, resultTag := range tokenResult.Tags {
		item.tags = append(item.tags, NewTag(resultTag))
	}

	err = store.Save(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (ic *ItemCreator) Edit(item *Item, data string, timeString string) error {
	store := ic.ctx.Value("store").(Store)
	// set a new time
	newAt := item.Time()
	if timeString != "" {
		span, err := TimeParser{Input: timeString}.Parse()
		if err != nil {
			return err
		}
		newAt = span.Start
	}
	err := ic.deleteOldItemTags(item)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to delete old item tags, continuing. %v\n", err)
		}
	}

	// set new description
	newDescription := item.Data()
	if data != "" {
		newDescription = data
	}

	item.datetime = newAt
	item.data = newDescription

	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(data)
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

	return store.Save(item)
}

func (ic *ItemCreator) Delete(item *Item) error {
	store := ic.ctx.Value("store").(Store)
	err := ic.deleteOldItemTags(item)
	if err != nil {
		if ic.isVerbose() {
			fmt.Printf("failed to delete old item tags. %v\n", err)
		}
		return err
	}
	err = store.Delete(item)
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
		err = store.SaveItemTag(item, tag)
		if err != nil {
			return err
		}
		if ic.isVerbose() {
			fmt.Printf("saved tag %s\n", tag.Name())
		}
	}
	return nil
}

func (ic *ItemCreator) deleteOldItemTags(item *Item) error {
	store := ic.ctx.Value("store").(Store)
	return store.DeleteItemTagsWithItem(item)
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