package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/josler/wdid/fileedit"
)

// edit description and time, not status
func Edit(ctx context.Context, idString string, description io.Reader, timeString string) error {
	store := ctx.Value("store").(Store)
	items, err := store.FindAll(idString)
	if err != nil {
		return err
	}
	if len(items) > 1 {
		printFormat := GetPrintFormatFromContext(ctx)
		if printFormat == HumanPrintFormat {
			fmt.Println("Error: Found multiple matching items:")
			NewItemPrinter(ctx).Print(items...)
		}
		return errors.New("unable to find unique item")
	}

	item := items[0]

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
	store := ctx.Value("store").(Store)
	items, err := store.FindAll(editID)
	if err != nil {
		return err
	}
	if len(items) > 1 {
		printFormat := GetPrintFormatFromContext(ctx)
		if printFormat == HumanPrintFormat {
			fmt.Println("Error: Found multiple matching items:")
			NewItemPrinter(ctx).Print(items...)
		}
		return errors.New("unable to find unique item")
	}

	data, err := fileedit.EditExisting(items[0].Data())
	if err != nil {
		return err
	}
	return Edit(ctx, editID, data, "")
}
