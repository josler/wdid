package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type StormItem struct {
	RowID      uint64 `storm:"id,increment"`
	ID         string `storm:"index,unique"`
	NextID     string
	PreviousID string
	Data       string
	Status     string
	Datetime   int64 `storm:"index"`
}

type BoltStore struct {
	db      *storm.DB
	ctx     context.Context
	itemMap map[string]*Item
}

func NewBoltStore(db *storm.DB) *BoltStore {
	return &BoltStore{db: db}
}

func (s *BoltStore) Find(id string) (*Item, error) {
	items, err := s.FindAll(id)
	if err != nil {
		return nil, err
	}
	if len(items) > 1 {
		return nil, errors.New("unable to find unique item")
	}
	return items[0], nil
}

func (s *BoltStore) FindAll(id string) ([]*Item, error) {
	stormItems := []*StormItem{}
	err := s.db.Prefix("ID", id, &stormItems)
	if err != nil {
		return nil, err
	}
	if len(stormItems) < 1 {
		return nil, errors.New("not found")
	}

	items := []*Item{}
	for _, stormItem := range stormItems {
		item, err := s.stormToItem(stormItem)
		if err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *BoltStore) Delete(item *Item) error {
	stormItem := s.itemToNewStorm(item)
	i, err := strconv.ParseUint(item.internalID, 10, 64)
	if err != nil {
		return nil
	}
	stormItem.RowID = i
	return s.db.DeleteStruct(stormItem)
}

func (s *BoltStore) Save(item *Item) error {
	stormItem := s.itemToNewStorm(item)
	if item.internalID != "" {
		i, err := strconv.ParseUint(item.internalID, 10, 64)
		if err != nil {
			return err
		}
		stormItem.RowID = i
		return s.db.Update(stormItem)
	}
	err := s.db.Save(stormItem)
	if err != nil {
		return err
	}
	item.internalID = fmt.Sprintf("%d", stormItem.RowID)
	return nil
}

func (s *BoltStore) List(t *Timespan, statuses ...string) ([]*Item, error) {
	stormItems := []*StormItem{}
	var query storm.Query
	if len(statuses) > 0 {
		query = s.db.Select(q.Gte("Datetime", t.Start.Unix()), q.Lte("Datetime", t.End.Unix()), q.In("Status", statuses))
	} else {
		query = s.db.Select(q.Gte("Datetime", t.Start.Unix()), q.Lte("Datetime", t.End.Unix()))
	}

	query.OrderBy("Datetime")
	err := query.Find(&stormItems)

	outputItems := []*Item{}
	if err != nil {
		if err == storm.ErrNotFound {
			return outputItems, nil
		}
		return outputItems, err
	}

	for _, item := range stormItems {
		parsed, err := s.stormToItem(item)
		if err != nil {
			return outputItems, err
		}
		outputItems = append(outputItems, parsed)
	}
	return outputItems, nil
}

func (s *BoltStore) WithContext(ctx context.Context) Store {
	return &BoltStore{ctx: ctx, db: s.db}
}

func (s *BoltStore) Close() {
	s.db.Close()
}

func (s *BoltStore) DropBucket(bucket string) {
	s.db.Drop(bucket)
}

func (s *BoltStore) itemToNewStorm(input *Item) *StormItem {
	return &StormItem{
		ID:         input.ID(),
		PreviousID: input.PreviousID(),
		NextID:     input.NextID(),
		Data:       input.Data(),
		Status:     input.Status(),
		Datetime:   input.Time().Unix(),
	}
}

func (s *BoltStore) stormToItem(input *StormItem) (*Item, error) {
	parsedTime := time.Unix(input.Datetime, 0)
	return &Item{
		internalID: fmt.Sprintf("%d", input.RowID),
		id:         input.ID,
		previousID: input.PreviousID,
		nextID:     input.NextID,
		data:       input.Data,
		status:     input.Status,
		datetime:   parsedTime,
	}, nil
}
