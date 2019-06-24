package core

import (
	"context"
	"errors"
	"fmt"
)

func Rm(ctx context.Context, idString string) error {
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

	NewItemPrinter(ctx).Print(items...)

	itemCreator := &ItemCreator{ctx: ctx}
	return itemCreator.Delete(items[0])
}
