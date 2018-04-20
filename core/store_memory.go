package core

import (
	"context"
	"errors"
	"strings"
	"time"
)

// MemoryStore for tests etc
type MemoryStore struct {
	ctx     context.Context
	itemMap map[string]*Item
}

func (s *MemoryStore) Find(id string) (*Item, error) {
	if len(id) < MAX_ID_LENGTH {
		for _, item := range s.itemMap {
			if strings.HasPrefix(item.ID(), id) {
				return item, nil
			}
		}
		return nil, errors.New("item not found")
	}

	found, ok := s.itemMap[id]
	if !ok {
		return nil, errors.New("item not found")
	}
	return found, nil
}

func (s *MemoryStore) Delete(item *Item) error {
	s.itemMap[item.ID()] = nil
	return nil
}

func (s *MemoryStore) Save(item *Item) error {
	if item.internalID == "" {
		item.internalID = item.ID()
	}
	s.itemMap[item.ID()] = item
	return nil
}

func (s *MemoryStore) List(t time.Time, statuses ...string) ([]*Item, error) {
	items := []*Item{}
	for _, item := range s.itemMap {
		if item.Time().Before(t) {
			continue // if before the time, skip
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

func (s *MemoryStore) WithContext(ctx context.Context) Store {
	return &MemoryStore{ctx: ctx}
}

func (s *MemoryStore) Close() {
}
