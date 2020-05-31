package core

import (
	"context"
	"errors"
	"fmt"
)

func FindAll(ctx context.Context, idString string) ([]*Item, error) {
	store := ctx.Value("store").(Store)
	return store.FindAll(idString)
}

func FindOneOrPrint(ctx context.Context, idString string) (*Item, error) {
	store := ctx.Value("store").(Store)
	items, err := store.FindAll(idString)
	if err != nil {
		return nil, err
	}
	if len(items) > 1 {
		printFormat := GetPrintFormatFromContext(ctx)
		if printFormat == HumanPrintFormat {
			fmt.Println("Error: Found multiple matching items:")
			NewItemPrinter(ctx).Print(items...)
		}
		return nil, errors.New("unable to find unique item")
	}
	return items[0], nil
}
