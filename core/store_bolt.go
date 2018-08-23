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

type StormTag struct {
	RowID     uint64 `storm:"id,increment"`
	Name      string `storm:"index,unique"`
	CreatedAt int64  // timestamp
	Type      string `storm:"index"`
}

type StormItemTag struct {
	RowID        uint64 `storm:"id,increment"`
	ItemTagIndex string `storm:"index,unique"`
	ItemID       string `storm:"index"`
	TagID        string `storm:"index"`
	CreatedAt    int64  `storm:"index"` // timestamp
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

func (s *BoltStore) FindTag(name string) (*Tag, error) {
	stormTag := &StormTag{}
	err := s.db.One("Name", name, stormTag)
	if err != nil {
		return nil, err
	}
	return s.stormToTag(stormTag)
}

func (s *BoltStore) SaveTag(tag *Tag) error {
	stormTag := s.tagToStorm(tag)
	err := s.db.Save(stormTag)
	if err != nil {
		if err == storm.ErrAlreadyExists {
			found, _ := s.FindTag(tag.Name())
			tag.internalID = found.internalID
			return nil
		}
		return err
	}
	tag.internalID = fmt.Sprintf("%d", stormTag.RowID)
	return nil
}

func (s *BoltStore) ListTags() ([]*Tag, error) {
	stormTags := []*StormTag{}
	query := s.db.Select()
	query.OrderBy("CreatedAt")
	err := query.Find(&stormTags)

	outputTags := []*Tag{}
	if err != nil {
		if err == storm.ErrNotFound {
			return outputTags, nil
		}
		return outputTags, err
	}

	for _, tag := range stormTags {
		parsed, err := s.stormToTag(tag)
		if err != nil {
			return outputTags, err
		}
		outputTags = append(outputTags, parsed)
	}
	return outputTags, nil
}

func (s *BoltStore) SaveItemTag(item *Item, tag *Tag) error {
	itemTag := NewItemTag(item, tag)
	stormItemTag := s.itemTagToStorm(itemTag)
	err := s.db.Save(stormItemTag)
	if err != nil {
		if err == storm.ErrAlreadyExists {
			// already exists
			return nil
		}
		return err
	}
	return nil
}

func (s *BoltStore) DeleteItemTag(item *Item, tag *Tag) error {
	itemTag := NewItemTag(item, tag)

	stormItemTags := []*StormItemTag{}
	query := s.db.Select(q.Eq("ItemTagIndex", fmt.Sprintf("%s:%s", itemTag.ItemID(), itemTag.TagID())))
	query.OrderBy("CreatedAt")
	err := query.Find(&stormItemTags)
	if err != nil {
		return err
	}
	err = s.db.DeleteStruct(stormItemTags[0])
	if err != nil {
		return err
	}
	return nil
}

func (s *BoltStore) FindItemsWithTag(tag *Tag) ([]*Item, error) {
	stormItemTags := []*StormItemTag{}
	query := s.db.Select(q.Eq("TagID", tag.internalID))
	query.OrderBy("CreatedAt").Limit(100).Reverse()
	err := query.Find(&stormItemTags)
	if err != nil {
		if err == storm.ErrNotFound {
			return []*Item{}, nil
		}
		return []*Item{}, err
	}

	stormItems := []*StormItem{}
	matchers := []q.Matcher{}
	for _, stormItemTag := range stormItemTags {
		matchers = append(matchers, q.Eq("ID", stormItemTag.ItemID))
	}
	query = s.db.Select(q.Or(matchers...))
	err = query.Find(&stormItems)

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

func (s *BoltStore) DeleteItemTagsWithItem(item *Item) error {
	stormItemTags := []*StormItemTag{}
	query := s.db.Select(q.Eq("ItemID", item.ID()))
	query.OrderBy("CreatedAt")
	err := query.Find(&stormItemTags)
	if err != nil {
		if err == storm.ErrNotFound {
			// nothing to do
			return nil
		}
		return err
	}
	for _, stormItemTag := range stormItemTags {
		err = s.db.DeleteStruct(stormItemTag)
		if err != nil {
			return err
		}
	}

	return nil
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

func (s *BoltStore) tagToStorm(input *Tag) *StormTag {
	return &StormTag{
		Name:      input.Name(),
		CreatedAt: input.CreatedAt().Unix(),
		Type:      input.TagType(),
	}
}

func (s *BoltStore) stormToTag(input *StormTag) (*Tag, error) {
	parsedTime := time.Unix(input.CreatedAt, 0)
	return &Tag{
		internalID: fmt.Sprintf("%d", input.RowID),
		name:       input.Name,
		createdAt:  parsedTime,
	}, nil
}

func (s *BoltStore) itemTagToStorm(input *ItemTag) *StormItemTag {
	return &StormItemTag{
		ItemTagIndex: fmt.Sprintf("%s:%s", input.ItemID(), input.TagID()),
		ItemID:       input.ItemID(),
		TagID:        input.TagID(),
		CreatedAt:    input.CreatedAt().Unix(),
	}
}

func (s *BoltStore) stormToItemTag(input *StormItemTag) (*ItemTag, error) {
	parsedTime := time.Unix(input.CreatedAt, 0)
	return &ItemTag{
		itemID:    input.ItemID,
		tagID:     input.TagID,
		createdAt: parsedTime,
	}, nil
}
