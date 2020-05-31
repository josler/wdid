package core

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/josler/wdid/fileedit"
)

// Edit description and time, not status
func Edit(ctx context.Context, idString string, description io.Reader, timeString string) error {
	item, err := FindOneOrPrint(ctx, idString)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	stringDescription = strings.Trim(stringDescription, "\n")

	itemCreator := &ItemCreator{ctx: ctx}
	item, err = itemCreator.Edit(item, stringDescription, timeString)
	if err != nil {
		return err
	}
	NewItemPrinter(ctx).Print(item)
	return nil
}

func EditDataFromFile(ctx context.Context, editID string) error {
	// find the item in question
	item, err := FindOneOrPrint(ctx, editID)
	if err != nil {
		return err
	}
	data, err := fileedit.EditExisting(item.Data())
	if err != nil {
		return err
	}
	return Edit(ctx, editID, data, "")
}
