package core

import (
	"bytes"
	"context"
	"io"
)

// edit description and time, not status
func Edit(ctx context.Context, idString string, description io.Reader, timeString string) error {
	store := ctx.Value("store").(Store)
	item, err := store.Find(idString)
	if err != nil {
		return err
	}

	// set a new time
	newAt := item.Time()
	if timeString != "" {
		span, err := TimeParser{Input: timeString}.Parse()
		if err != nil {
			return err
		}
		newAt = span.Start
	}

	// set new description
	newDescription := item.Data()
	buf := new(bytes.Buffer)
	buf.ReadFrom(description)
	stringDescription := buf.String()
	if stringDescription != "" {
		newDescription = stringDescription
	}

	item.datetime = newAt
	item.data = newDescription
	return store.Save(item)
}
