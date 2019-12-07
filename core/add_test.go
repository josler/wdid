package core

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/josler/wdid/filter"
	"gotest.tools/assert"
)

func TestAddCreatesItem(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item #hashtag"), "now")
		found := mostRecentItem(store)
		if found.Data() != "my new item #hashtag" {
			t.Errorf("item not saved")
		}
		tag, err := store.FindTag("#hashtag")
		if err != nil || tag.Name() != "#hashtag" {
			t.Errorf("tag not saved")
		}
	})
}

func TestAddFutureItem(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		err := Add(ctx, strings.NewReader("my new item"), "2025-01-01")
		if err != nil {
			t.Errorf("item not saved")
		}
	})
}

func TestAddDone(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		err := AddDone(ctx, strings.NewReader("my new item"), "now")
		if err != nil {
			t.Errorf("item not saved")
		}
		item := mostRecentItem(store)
		assert.Equal(t, item.Status(), DoneStatus)
	})
}

func contextWithStore(f func(ctx context.Context, store Store)) {
	ctx := context.Background()

	store, err := NewBoltStore("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}

	store.DropBucket("StormItem")
	store.DropBucket("StormTag")
	store.DropBucket("StormGroup")

	ctx = context.WithValue(ctx, "store", store)
	f(ctx, store)
}

// mostRecentItem returns the most recent item chronologically in the store, up until time.Now()
// it does not deal with insertion order
func mostRecentItem(store Store) *Item {
	filters := []filter.Filter{NewDateFilter(filter.FilterEq, NewTimespan(time.Unix(0, 0), time.Now()))}
	items, err := store.ListFilters(filters)
	if err != nil {
		panic(err)
	}
	return items[len(items)-1]
}
