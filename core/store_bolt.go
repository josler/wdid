package core

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/asdine/storm"
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
	Kind       int64 `storm:"index"`
}

type StormTag struct {
	RowID     uint64 `storm:"id,increment"`
	Name      string `storm:"index,unique"`
	CreatedAt int64  // timestamp
	Type      string `storm:"index"`
}

type StormGroup struct {
	RowID        uint64 `storm:"id,increment"`
	Name         string `storm:"index,unique"`
	FilterString string
	CreatedAt    int64 `storm:"index"` // timestamp
}

type BoltStore struct {
	path string
	ctx  context.Context
}

func NewBoltStore(path string) (*BoltStore, error) {
	store := &BoltStore{path: path}
	db, err := storm.Open(store.path)
	if err != nil {
		return nil, err
	}
	err = db.Close()
	return store, err
}

func (s *BoltStore) withOpenDB(f func(*storm.DB)) {
	db, err := storm.Open(s.path)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	f(db)
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
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.Prefix("ID", id, &stormItems)
	})

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
	s.withOpenDB(func(db *storm.DB) {
		err = db.DeleteStruct(stormItem)
	})
	return err
}

func (s *BoltStore) Save(item *Item) error {
	stormItem := s.itemToNewStorm(item)
	if item.internalID != "" {
		i, err := strconv.ParseUint(item.internalID, 10, 64)
		if err != nil {
			return err
		}
		stormItem.RowID = i
		s.withOpenDB(func(db *storm.DB) {
			err = db.Update(stormItem)
		})
		return err
	}
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.Save(stormItem)
	})
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
	var err error

	s.withOpenDB(func(db *storm.DB) {
		if firstDateFilter != nil {
			// if we have a date filter, use it as a range to limit where we search over
			err = db.Range("Datetime", firstDateFilter.timespan.Start.Unix(), firstDateFilter.timespan.End.Unix(), &stormItems)
		} else {
			// else, get all
			err = db.All(&stormItems)
		}
	})

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

func (s *BoltStore) FindTag(name string) (*Tag, error) {
	stormTag := &StormTag{}
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.One("Name", name, stormTag)
	})

	if err != nil {
		return nil, err
	}
	return s.stormToTag(stormTag)
}

func (s *BoltStore) SaveTag(tag *Tag) error {
	stormTag := s.tagToStorm(tag)
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.Save(stormTag)
	})
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
	var err error
	s.withOpenDB(func(db *storm.DB) {
		query := db.Select()
		query.OrderBy("CreatedAt")
		err = query.Find(&stormTags)
	})

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

func (s *BoltStore) SaveGroup(group *Group) error {
	stormGroup := s.groupToStorm(group)
	if group.internalID != "" {
		i, err := strconv.ParseUint(group.internalID, 10, 64)
		if err != nil {
			return err
		}
		stormGroup.RowID = i
		s.withOpenDB(func(db *storm.DB) {
			err = db.Update(stormGroup)
		})
		return err
	}
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.Save(stormGroup)
	})
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
	s.withOpenDB(func(db *storm.DB) {
		err = db.DeleteStruct(stormGroup)
	})
	return err
}

func (s *BoltStore) ListGroups() ([]*Group, error) {
	stormGroups := []*StormGroup{}
	var err error
	s.withOpenDB(func(db *storm.DB) {
		query := db.Select()
		query.OrderBy("CreatedAt")
		err = query.Find(&stormGroups)
	})

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
	var err error
	s.withOpenDB(func(db *storm.DB) {
		err = db.One("Name", name, stormGroup)
	})
	if err != nil {
		return nil, err
	}
	return s.stormToGroup(stormGroup)
}

func (s *BoltStore) WithContext(ctx context.Context) Store {
	return &BoltStore{ctx: ctx, path: s.path}
}

func (s *BoltStore) DropBucket(bucket string) {
	s.withOpenDB(func(db *storm.DB) {
		db.Drop(bucket)
	})
}

func (s *BoltStore) itemToNewStorm(input *Item) *StormItem {
	return &StormItem{
		ID:         input.ID(),
		PreviousID: input.PreviousID(),
		NextID:     input.NextID(),
		Data:       input.Data(),
		Status:     input.Status(),
		Datetime:   input.Time().Unix(),
		Kind:       int64(input.Kind()),
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
		kind:       Kind(input.Kind),
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
