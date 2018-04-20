package core

import (
	"context"
	"errors"
	"time"
)

func Bump(ctx context.Context, id string, timeString string) error {
	store := ctx.Value("store").(Store)
	item, err := store.WithContext(ctx).Find(id)
	if err != nil {
		return err
	}
	if item.Status() != "waiting" {
		return errors.New("can't bump finished item")
	}

	to, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}
	if to.After(time.Now()) {
		return errors.New("can't set future time")
	}

	newItem := item.Bump(to) // mark old item as done

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
