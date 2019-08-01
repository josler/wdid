package core

import "context"

func FindAll(ctx context.Context, idString string) ([]*Item, error) {
	store := ctx.Value("store").(Store)
	return store.FindAll(idString)
}
