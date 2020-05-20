package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func Import(ctx context.Context, filename string) error {
	var f io.Reader
	fmt.Println(filename)

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		f = file
	} else {
		f = os.Stdin
	}
	return ReadToStore(ctx, f)
}

func ReadToStore(ctx context.Context, f io.Reader) error {
	store := ctx.Value("store").(Store)
	items := []*Item{}
	itemCreator := &ItemCreator{ctx: ctx}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data := scanner.Text()
		split := strings.Split(data, "\t")

		parsedTime, err := time.Parse(time.RFC3339, split[5])
		if err != nil {
			continue
		}
		var kind Kind
		if len(split) == 7 {
			kind = StringToKind(split[6])
		} else {
			kind = Task
		}

		item := &Item{id: split[0], internalID: split[1], status: split[2], data: split[4], datetime: parsedTime, kind: kind}

		refID := split[3]
		if strings.HasPrefix(refID, "->") {
			item.nextID = refID[2:]
		} else if strings.HasPrefix(refID, "<-") {
			item.previousID = refID[2:]
		}

		// not worrying about errors on metadata
		itemCreator.GenerateAndSaveMetadata(item)
		items = append(items, item)
	}

	for _, item := range items {
		found, err := store.Find(item.ID())
		if err != nil { // not found, issue regular save
			item.ResetInternalID() // we want to re-issue this
		} else { // we have a match on the ID
			if found.internalID != item.internalID { // internal id doesn't match! use the existing one
				item.internalID = found.internalID
			}
		}
		err = store.Save(item)

		if err != nil {
			return err
		}
	}
	return nil
}
