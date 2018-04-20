package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"
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
		newAt, err = TimeParser{Input: timeString}.Parse()
		if err != nil {
			return err
		}
		if newAt.After(time.Now()) {
			return errors.New("can't set future time")
		}
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
