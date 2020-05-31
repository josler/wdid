package core

import (
	"context"
)

func Do(ctx context.Context, id string) error {
	item, err := FindOneOrPrint(ctx, id)
	if err != nil {
		return err
	}

	item.Do()
	store := ctx.Value("store").(Store)
	err = store.WithContext(ctx).Save(item)
	NewItemPrinter(ctx).Print(item)
	return err
}
