package core

import (
	"context"

	"github.com/blevesearch/bleve"
)

func Skip(ctx context.Context, id string) error {
	store := ctx.Value("store").(Store)
	item, err := store.WithContext(ctx).Find(id)
	if err != nil {
		return err
	}
	item.Skip()
	err = store.WithContext(ctx).Save(item)
	if err != nil {
		return err
	}
	index := ctx.Value("index").(bleve.Index)
	SaveBleve(index, store, item)
	NewItemPrinter(ctx).Print(item)
	return err
}
