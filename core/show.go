package core

import (
	"context"
)

func Show(ctx context.Context, idString string) error {
	store := ctx.Value("store").(Store)
	item, err := store.Find(idString)
	if err != nil {
		return err
	}
	NewItemPrinter(ctx).Print(item)
	return nil
}
