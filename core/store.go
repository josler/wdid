package core

import (
	"context"
)

const (
	MAX_ID_LENGTH = 6
	SCAN_LIMIT    = 14
)

type Store interface {
	ItemStore
	TagStore
	ItemTagStore

	WithContext(ctx context.Context) Store
	Close()
}

type ItemStore interface {
	Find(id string) (*Item, error)
	FindAll(id string) ([]*Item, error)
	Delete(item *Item) error
	Save(item *Item) error
	List(t *Timespan, statuses ...string) ([]*Item, error)
}

type TagStore interface {
	FindTag(name string) (*Tag, error)
	SaveTag(tag *Tag) error
	ListTags() ([]*Tag, error)
}

type ItemTagStore interface {
	SaveItemTag(item *Item, tag *Tag) error
	DeleteItemTag(item *Item, Tag *Tag) error
	FindItemsWithTag(tag *Tag, limit int) ([]*Item, error)
	DeleteItemTagsWithItem(item *Item) error
}
