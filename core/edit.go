package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
		if printFormat == HUMAN_PRINT_FORMAT {
			fmt.Println("Error: Found multiple matching items:")
			NewItemPrinter(ctx).Print(items...)
		}
		return errors.New("unable to find unique item")
	}

	item := items[0]

	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()

	itemCreator := &ItemCreator{ctx: ctx}
	item, err = itemCreator.Edit(item, stringDescription, timeString)
	NewItemPrinter(ctx).Print(item)
	return err
}
