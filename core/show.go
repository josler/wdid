package core

import (
	"context"
)

func Show(ctx context.Context, idString string) error {
	store := ctx.Value("store").(Store)
	items, err := store.FindAll(idString)
	if err != nil {
		return err
	}
	NewItemPrinter(ctx).Print(items...)
	return nil
}
