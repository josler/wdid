package core

import (
	"context"
)

const (
	MAX_ID_LENGTH = 6
	SCAN_LIMIT    = 14
)

type Store interface {
	Find(id string) (*Item, error)
	FindAll(id string) ([]*Item, error)
	Delete(item *Item) error
	Save(item *Item) error
	List(t *Timespan, statuses ...string) ([]*Item, error)
	WithContext(ctx context.Context) Store
	Close()
}
