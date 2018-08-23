package core

import (
	"bytes"
	"context"
	"errors"
	"io"
)

func Add(ctx context.Context, description io.Reader, timeString string) error {
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
	itemCreator := &ItemCreator{ctx: ctx}
	item, err := itemCreator.Create(stringDescription, at.Start)
	if err != nil {
		return err
	}

	NewItemPrinter(ctx).Print(item)
	return nil
}
