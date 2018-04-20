package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAddCreatesItem(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "now")
	found := mostRecentItem(store)
	if found.Data() != "my new item" {
		t.Errorf("item not saved")
	}
}

func TestAddPreventsFutureItem(t *testing.T) {
	ctx, _ := contextWithMemoryStore()
	err := Add(ctx, strings.NewReader("my new item"), "2025-01-01")
	if err == nil {
		t.Errorf("item was incorrectly saved")
	}
}

func contextWithMemoryStore() (context.Context, Store) {
	ctx := context.Background()
	store := &MemoryStore{itemMap: map[string]*Item{}}
	return context.WithValue(ctx, "store", store), store
}

func mostRecentItem(store Store) *Item {
	items, err := store.List(time.Unix(0, 0))
	if err != nil {
		panic(err)
	}
	return items[len(items)-1]
}
