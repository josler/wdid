package core

import (
	"context"
	"errors"

	"github.com/josler/wdid/filter"
)

var ErrorNotYetImplemented error = errors.New("Not yet implemented")

type SqlStore struct {
	ctx context.Context
}

func NewSqlStore() *SqlStore {
	return &SqlStore{}
}

func (store *SqlStore) WithContext(ctx context.Context) Store {
	return &SqlStore{ctx: ctx}
}

func (store *SqlStore) Close() {
}

func (store *SqlStore) Find(id string) (*Item, error) {
	return nil, ErrorNotYetImplemented
}

func (store *SqlStore) FindAll(id string) ([]*Item, error) {
	return []*Item{}, ErrorNotYetImplemented
}

func (store *SqlStore) Delete(item *Item) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) Save(item *Item) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) List(t *Timespan, statuses ...string) ([]*Item, error) {
	return []*Item{}, ErrorNotYetImplemented
}

func (store *SqlStore) ListFilters(filters []filter.Filter) ([]*Item, error) {
	return []*Item{}, ErrorNotYetImplemented
}

func (store *SqlStore) FindTag(name string) (*Tag, error) {
	return nil, ErrorNotYetImplemented
}

func (store *SqlStore) SaveTag(tag *Tag) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) ListTags() ([]*Tag, error) {
	return []*Tag{}, ErrorNotYetImplemented
}

func (store *SqlStore) SaveItemTag(item *Item, tag *Tag) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) DeleteItemTag(item *Item, Tag *Tag) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) FindItemsWithTag(tag *Tag, limit int) ([]*Item, error) {
	return []*Item{}, ErrorNotYetImplemented
}

func (store *SqlStore) DeleteItemTagsWithItem(item *Item) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) SaveGroup(group *Group) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) DeleteGroup(group *Group) error {
	return ErrorNotYetImplemented
}

func (store *SqlStore) ListGroups() ([]*Group, error) {
	return []*Group{}, ErrorNotYetImplemented
}

func (store *SqlStore) FindGroupByName(name string) (*Group, error) {
	return nil, ErrorNotYetImplemented
}
