package core

import (
	"context"
	"time"
)

const (
	MAX_ID_LENGTH = 6
	SCAN_LIMIT    = 14
)

type Store interface {
	Find(id string) (*Item, error)
	Delete(item *Item) error
	Save(item *Item) error
	List(t time.Time, statuses ...string) ([]*Item, error)
	WithContext(ctx context.Context) Store
	Close()
}
