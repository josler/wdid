package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// MemoryStore for tests etc
type MemoryStore struct {
	ctx        context.Context
	itemMap    map[string]*Item
	itemTagMap map[string]*ItemTag
	tagMap     map[string]*Tag
}

func (s *MemoryStore) Find(id string) (*Item, error) {
	items, err := s.FindAll(id)
	if err != nil {
		return nil, err
	}
	if len(items) > 1 {
		return nil, errors.New("unable to find unique item")
	}
	return items[0], nil
}

func (s *MemoryStore) FindAll(id string) ([]*Item, error) {
	// full match
	if len(id) == MAX_ID_LENGTH {
		found, ok := s.itemMap[id]
		if !ok {
			return []*Item{}, errors.New("item not found")
		}
		return []*Item{found}, nil
	}

	// partial match
	items := []*Item{}
	for _, item := range s.itemMap {
		if strings.HasPrefix(item.ID(), id) {
			items = append(items, item)
		}
	}
	if len(items) < 1 {
		return nil, errors.New("item not found")
	}
	return items, nil
}

func (s *MemoryStore) Delete(item *Item) error {
	delete(s.itemMap, item.ID())
	return nil
}

func (s *MemoryStore) Save(item *Item) error {
	if item.internalID == "" {
		item.internalID = item.ID()
	}
	s.itemMap[item.ID()] = item
	return nil
}

func (s *MemoryStore) List(t *Timespan, statuses ...string) ([]*Item, error) {
	items := []*Item{}
	for _, item := range s.itemMap {
		if item.Time().Before(t.Start) {
			continue // if before the time, skip
		}

		if item.Time().After(t.End) {
			continue // if after, skip
		}

		if len(statuses) > 0 && !s.includes(item.Status(), statuses...) {
			continue // if it has statuses, then skip if not in one of them
		}

		items = append(items, item)
	}
	return items, nil
}

func (s *MemoryStore) includes(status string, statuses ...string) bool {
	for _, toMatch := range statuses {
		if status == toMatch {
			return true
		}
	}
	return false
}

func (s *MemoryStore) FindTag(name string) (*Tag, error) {
	found, ok := s.tagMap[name]
	if !ok {
		return nil, errors.New("item not found")
	}
	return found, nil
}

func (s *MemoryStore) SaveTag(tag *Tag) error {
	s.tagMap[tag.Name()] = tag
	return nil
}

func (s *MemoryStore) ListTags() ([]*Tag, error) {
	tagList := []*Tag{}
	for _, tag := range s.tagMap {
		tagList = append(tagList, tag)
	}
	return tagList, nil
}

func (s *MemoryStore) SaveItemTag(item *Item, tag *Tag) error {
	itemTag := NewItemTag(item, tag)
	id := fmt.Sprintf("%s:%s", itemTag.TagID(), itemTag.ItemID())
	s.itemTagMap[id] = itemTag
	return nil
}

func (s *MemoryStore) DeleteItemTag(item *Item, tag *Tag) error {
	itemTag := NewItemTag(item, tag)
	id := fmt.Sprintf("%s:%s", itemTag.TagID(), itemTag.ItemID())
	delete(s.itemTagMap, id)
	return nil
}

func (s *MemoryStore) FindItemsWithTag(tag *Tag) ([]*Item, error) {
	items := []*Item{}
	for k, itemTag := range s.itemTagMap {
		if strings.HasPrefix(k, tag.internalID) {
			item := s.itemMap[itemTag.ItemID()]
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *MemoryStore) WithContext(ctx context.Context) Store {
	return &MemoryStore{ctx: ctx}
}

func (s *MemoryStore) Close() {
}
