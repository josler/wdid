package core

import (
	"context"
	"errors"
	"fmt"
)

func Do(ctx context.Context, id string) error {
	store := ctx.Value("store").(Store)
	items, err := store.FindAll(id)
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
	item.Do()
	err = store.WithContext(ctx).Save(item)
	NewItemPrinter(ctx).Print(item)
	return err
}
