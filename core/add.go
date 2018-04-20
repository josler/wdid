package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"
)

func Add(ctx context.Context, description io.Reader, timeString string) error {
	store := ctx.Value("store").(Store)
	at, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return err
	}
	if at.After(time.Now()) {
		return errors.New("can't set future time")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	if stringDescription == "" {
		return errors.New("description missing")
	}
	item := NewItem(stringDescription, at)
	err = store.Save(item)
	if err == nil {
		NewItemPrinter(ctx).Print(item)
	}
	return err
}
