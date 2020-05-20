package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"
)

func Add(ctx context.Context, description io.Reader, timeString string) error {
	itemCreator := &ItemCreator{ctx: ctx}
	item, err := addCreate(description, timeString, itemCreator.CreateTask)
	if err != nil {
		return err
	}

	NewItemPrinter(ctx).Print(item)
	return nil
}

func AddDone(ctx context.Context, description io.Reader, timeString string) error {
	itemCreator := &ItemCreator{ctx: ctx}
	item, err := addCreate(description, timeString, itemCreator.CreateTask)
	if err != nil {
		return err
	}
	Do(ctx, item.ID())
	return nil
}

func AddNote(ctx context.Context, description io.Reader, timeString string) error {
	itemCreator := &ItemCreator{ctx: ctx}
	item, err := addCreate(description, timeString, itemCreator.CreateNote)
	if err != nil {
		return err
	}
	connectedItems := getValidConnections(ctx, item)
	NewItemPrinter(ctx).PrintSingleWithConnected(item, connectedItems...)
	return nil
}

type itemCreateFn func(data string, at time.Time) (*Item, error)

func addCreate(description io.Reader, timeString string, creator itemCreateFn) (*Item, error) {
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
	return creator(stringDescription, at.Start)
}
