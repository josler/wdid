package core

import (
	"context"
)

func Rm(ctx context.Context, idString string) error {
	item, err := FindOneOrPrint(ctx, idString)
	if err != nil {
		return err
	}
	NewItemPrinter(ctx).Print(item)

	itemCreator := &ItemCreator{ctx: ctx}
	return itemCreator.Delete(item)
}
