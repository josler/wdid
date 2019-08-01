package core

import (
	"context"
)

func Show(ctx context.Context, idString string) error {
	items, err := FindAll(ctx, idString)
	if err != nil {
		return err
	}
	NewItemPrinter(ctx).Print(items...)
	return nil
}
