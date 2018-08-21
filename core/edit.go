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

	// set a new time
	newAt := item.Time()
	if timeString != "" {
		span, err := TimeParser{Input: timeString}.Parse()
		if err != nil {
			return err
		}
		newAt = span.Start
	}

	// set new description
	newDescription := item.Data()
	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	if stringDescription != "" {
		newDescription = stringDescription
	}

	item.datetime = newAt
	item.data = newDescription
	return store.Save(item)
}
