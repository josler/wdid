package core

import (
	"context"
	"errors"
)

func Bump(ctx context.Context, id string, timeString string) error {
	store := ctx.Value("store").(Store)
	item, err := FindOneOrPrint(ctx, id)
	if err != nil {
		return err
	}
	if item.Status() != WaitingStatus {
		return errors.New("can't bump finished item")
	}

	to, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}

	newItem := item.Bump(to.Start) // mark old item as done

	// save old
	err = store.WithContext(ctx).Save(item)
	if err != nil {
		return err
	}

	// save new
	err = store.WithContext(ctx).Save(newItem)
	NewItemPrinter(ctx).Print(newItem)
	return err
}
