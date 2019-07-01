package core

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/josler/wdid/filter"
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

type StormGroup struct {
	RowID        uint64 `storm:"id,increment"`
	Name         string `storm:"index,unique"`
	FilterString string
	CreatedAt    int64 `storm:"index"` // timestamp
}

type BoltStore struct {
	db  *storm.DB
	ctx context.Context
}

func NewBoltStore(db *storm.DB) *BoltStore {
	return &BoltStore{db: db}
}

func (s *BoltStore) Find(id string) (*Item, error) {
	items, err := s.FindAll(id)
	if err != nil {
		return nil, err
	}

	// if there's no items, then we failed
	if len(items) == 0 {
		return nil, errors.New("not found")
	}

	// return most recent item
	sort.Slice(items, func(i, j int) bool {
		return items[i].Time().After(items[j].Time())
	})

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

func (s *BoltStore) findFirstDateFilter(filters []filter.Filter) (*DateFilter, []filter.Filter) {
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

func (s *BoltStore) ListFilters(filters []filter.Filter) ([]*Item, error) {
	stormItems := []*StormItem{}
	outputItems := []*Item{}

	firstDateFilter, rest := s.findFirstDateFilter(filters)
	if firstDateFilter == nil {
		t, _ := TimeParser{Input: "0"}.Parse()
		firstDateFilter = NewDateFilter(filter.FilterEq, t)
	}

	err := s.db.Range("Datetime", firstDateFilter.timespan.Start.Unix(), firstDateFilter.timespan.End.Unix(), &stormItems)
	if err != nil {
		if err == storm.ErrNotFound {
			return outputItems, nil
		}
		return nil, err
	}

	for _, stormItem := range stormItems {
		match := true
		for _, filter := range rest {
			ok, err := filter.Match(*stormItem)
			if !ok || err != nil {
				match = false
				break
			}
		}

		if match {
			parsed, err := s.stormToItem(stormItem)
			if err != nil {
				return outputItems, err
			}
			outputItems = append(outputItems, parsed)
		}
	}
	sort.Slice(outputItems, func(i, j int) bool {
		return outputItems[i].Time().Before(outputItems[j].Time())
	})

	return outputItems, nil
}

func (s *BoltStore) List(t *Timespan, statuses ...string) ([]*Item, error) {
	filters := []filter.Filter{NewDateFilter(filter.FilterEq, t)}
	if len(statuses) > 0 {
		filters = append(filters, NewStatusFilter(filter.FilterEq, statuses...))
	}
	return s.ListFilters(filters)
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

func (s *BoltStore) FindItemsWithTag(tag *Tag, limit int) ([]*Item, error) {
	stormItemTags := []*StormItemTag{}
	query := s.db.Select(q.Eq("TagID", tag.internalID))
	if limit > 0 {
		query.OrderBy("CreatedAt").Limit(limit).Reverse()
	} else {
		query.OrderBy("CreatedAt").Reverse()
	}

	err := query.Find(&stormItemTags)
	if err != nil {
		if err == storm.ErrNotFound {
			return []*Item{}, nil
		}
		return []*Item{}, err
	}

	stormItems := []*StormItem{}
	ids := []string{}
	for _, stormItemTag := range stormItemTags {
		ids = append(ids, stormItemTag.ItemID)
	}
	query = s.db.Select(q.In("ID", ids))
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

func (s *BoltStore) SaveGroup(group *Group) error {
	stormGroup := s.groupToStorm(group)
	if group.internalID != "" {
		i, err := strconv.ParseUint(group.internalID, 10, 64)
		if err != nil {
			return err
		}
		stormGroup.RowID = i
		return s.db.Update(stormGroup)
	}
	err := s.db.Save(stormGroup)
	if err != nil {
		return err
	}
	group.internalID = fmt.Sprintf("%d", stormGroup.RowID)
	return nil
}

func (s *BoltStore) DeleteGroup(group *Group) error {
	stormGroup := s.groupToStorm(group)
	i, err := strconv.ParseUint(group.internalID, 10, 64)
	if err != nil {
		return nil
	}
	stormGroup.RowID = i
	return s.db.DeleteStruct(stormGroup)
}

func (s *BoltStore) ListGroups() ([]*Group, error) {
	stormGroups := []*StormGroup{}
	query := s.db.Select()
	query.OrderBy("CreatedAt")
	err := query.Find(&stormGroups)

	outputGroups := []*Group{}
	if err != nil {
		if err == storm.ErrNotFound {
			return outputGroups, nil
		}
		return outputGroups, err
	}

	for _, group := range stormGroups {
		parsed, err := s.stormToGroup(group)
		if err != nil {
			return outputGroups, err
		}
		outputGroups = append(outputGroups, parsed)
	}
	return outputGroups, nil
}

func (s *BoltStore) FindGroupByName(name string) (*Group, error) {
	stormGroup := &StormGroup{}
	err := s.db.One("Name", name, stormGroup)
	if err != nil {
		return nil, err
	}
	return s.stormToGroup(stormGroup)
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

func (s *BoltStore) groupToStorm(input *Group) *StormGroup {
	return &StormGroup{
		Name:         input.Name,
		FilterString: input.FilterString,
		CreatedAt:    input.CreatedAt.Unix(),
	}
}

func (s *BoltStore) stormToGroup(input *StormGroup) (*Group, error) {
	parsedTime := time.Unix(input.CreatedAt, 0)
	return &Group{
		internalID:   fmt.Sprintf("%d", input.RowID),
		Name:         input.Name,
		FilterString: input.FilterString,
		CreatedAt:    parsedTime,
	}, nil
}
