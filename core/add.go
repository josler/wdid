package core

import (
	"bytes"
	"context"
	"errors"
	"io"
)

func Add(ctx context.Context, description io.Reader, timeString string) error {
	store := ctx.Value("store").(Store)
	at, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	if stringDescription == "" {
		return errors.New("description missing")
	}
	item := NewItem(stringDescription, at.Start)
	err = store.Save(item)
	if err == nil {
		NewItemPrinter(ctx).Print(item)
	}
	return err
}
