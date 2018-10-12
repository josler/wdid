package core

import "context"

func Do(ctx context.Context, id string) error {
	store := ctx.Value("store").(Store)
	item, err := store.WithContext(ctx).Find(id)
	if err != nil {
		return err
	}
	item.Do()
	err = store.WithContext(ctx).Save(item)
	NewItemPrinter(ctx).Print(item)
	return err
}
