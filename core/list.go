package core

import (
	"context"
)

func List(ctx context.Context, timeString string, statuses ...string) error {
	store := ctx.Value("store").(Store)
	from, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}
	items, err := store.List(from, statuses...)
	NewItemPrinter(ctx).Print(items...)
	return err
}
