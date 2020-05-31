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
	GroupStore

	WithContext(ctx context.Context) Store
}

type ItemStore interface {
	FindAll(id string) ([]*Item, error)
	Delete(item *Item) error
	Save(item *Item) error
	ListFilters(filters []filter.Filter) ([]*Item, error)
}

type TagStore interface {
	FindTag(name string) (*Tag, error)
	SaveTag(tag *Tag) error
	ListTags() ([]*Tag, error)
}

type GroupStore interface {
	SaveGroup(group *Group) error
	DeleteGroup(group *Group) error
	ListGroups() ([]*Group, error)
	FindGroupByName(name string) (*Group, error)
}
