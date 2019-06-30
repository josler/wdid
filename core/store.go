package core

import (
	"context"

	"github.com/josler/wdid/filter"
)

const (
	MaxIDLength = 6
)

type Store interface {
	ItemStore
	TagStore
	ItemTagStore
	GroupStore

	WithContext(ctx context.Context) Store
	Close()
}

type ItemStore interface {
	Find(id string) (*Item, error)
	FindAll(id string) ([]*Item, error)
	Delete(item *Item) error
	Save(item *Item) error
	List(t *Timespan, statuses ...string) ([]*Item, error)
	ListFilters(filters []filter.Filter) ([]*Item, error)
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

type GroupStore interface {
	SaveGroup(group *Group) error
	FindGroupByName(name string) (*Group, error)
}
