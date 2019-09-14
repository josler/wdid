package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/josler/wdid/filter"
)

var ErrorNotYetImplemented error = errors.New("Not yet implemented")

type SqlStore struct {
	ctx context.Context
	db  *sqlx.DB
}

func NewSqlStore(db *sqlx.DB) *SqlStore {
	return &SqlStore{db: db}
}

func (store *SqlStore) WithContext(ctx context.Context) Store {
	return &SqlStore{ctx: ctx, db: store.db}
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

func (store *SqlStore) InitTables() error {
	var err error
	schema := `CREATE TABLE item (
		row_id INTEGER PRIMARY KEY AUTOINCREMENT,
		wdid_id TEXT UNIQUE,
		next_id TEXT NULL,
		previous_id TEXT NULL,
		data TEXT,
		status TEXT,
		datetime INTEGER);`
	if _, err = store.db.Exec(schema); err != nil {
		return err
	}

	item_index := `CREATE INDEX item_wdid_id ON item(wdid_id)`
	if _, err = store.db.Exec(item_index); err != nil {
		return err
	}

	date_index := `CREATE INDEX item_datetime ON item(datetime)`
	if _, err = store.db.Exec(date_index); err != nil {
		return err
	}
	return nil
}

func (store *SqlStore) Save(item *Item) error {
	itemSql := `INSERT INTO item (wdid_id, next_id, previous_id, data, status, datetime) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := store.db.Exec(itemSql, item.ID(), item.NextID(), item.PreviousID(), item.Data(), item.Status(), item.Time().Unix())
	return err
}

type SqlItem struct {
	RowID      int64  `db:"row_id"`
	ID         string `db:"wdid_id"`
	NextID     string `db:"next_id"`
	PreviousID string `db:"previous_id"`
	Data       string `db:"data"`
	Status     string `db:"status"`
	Datetime   int64  `db:"datetime"`
}

func (store *SqlStore) List(t *Timespan, statuses ...string) ([]*Item, error) {
	filters := []filter.Filter{NewDateFilter(filter.FilterEq, t)}
	if len(statuses) > 0 {
		filters = append(filters, NewStatusFilter(filter.FilterEq, statuses...))
	}
	return store.ListFilters(filters)
}

func (store *SqlStore) ListFilters(filters []filter.Filter) ([]*Item, error) {

	store.Save(NewItem("foo #foo", time.Now().Add(-24*time.Hour)))
	store.Save(NewItem("bar", time.Now()))

	sqlItems := []*SqlItem{}
	items := []*Item{}

	firstDateFilter, rest := store.findFirstDateFilter(filters)
	var err error

	if firstDateFilter != nil {
		// if we have a date filter, use it as a range to limit where we search over
		err = store.db.Select(&sqlItems, "SELECT * FROM item WHERE datetime BETWEEN ? AND ?", firstDateFilter.timespan.Start.Unix(), firstDateFilter.timespan.End.Unix())
	} else {
		// else, get all
		err = store.db.Select(&sqlItems, "SELECT * FROM item")
	}

	if err != nil {
		// TODO: check type of err
		return nil, err
	}

	for _, item := range sqlItems {
		match := true
		for _, filter := range rest {
			ok, err := filter.Match(*item)
			if !ok || err != nil {
				match = false
				break
			}
		}

		if match {
			// TODO: DRY up
			parsedTime := time.Unix(item.Datetime, 0)
			parsed := &Item{
				internalID: fmt.Sprintf("%d", item.RowID),
				id:         item.ID,
				nextID:     item.NextID,
				previousID: item.PreviousID,
				data:       item.Data,
				status:     item.Status,
				datetime:   parsedTime,
			}
			items = append(items, parsed)
		}
	}

	return items, nil
}

// TODO: move and DRY up
func (store *SqlStore) findFirstDateFilter(filters []filter.Filter) (*DateFilter, []filter.Filter) {
	var rest []filter.Filter
	for i, f := range filters {
		switch df := f.(type) {
		case *DateFilter:
			rest = append(filters[:i], filters[i+1:]...)
			return df, rest
		}
	}
	return nil, filters
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
