package core

import (
	"bytes"
	"context"
	"errors"
	"io"
)

func Add(ctx context.Context, description io.Reader, timeString string) error {
	item, err := addCreate(ctx, description, timeString)
	if err != nil {
		return err
	}

	NewItemPrinter(ctx).Print(item)
	return nil
}

func AddDone(ctx context.Context, description io.Reader, timeString string) error {
	item, err := addCreate(ctx, description, timeString)
	if err != nil {
		return err
	}
	Do(ctx, item.ID())
	return nil
}

func addCreate(ctx context.Context, description io.Reader, timeString string) (*Item, error) {
	at, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	if stringDescription == "" {
		return nil, errors.New("description missing")
	}
	itemCreator := &ItemCreator{ctx: ctx}
	item, err := itemCreator.Create(stringDescription, at.Start)
	if err != nil {
		return nil, err
	}
	return item, nil
}
