package core

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm"
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

func contextWithStore(f func(ctx context.Context, store Store)) {
	ctx := context.Background()

	db, err := storm.Open("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}

	store := NewBoltStore(db)
	store.DropBucket("StormItem")
	store.DropBucket("StormTag")
	store.DropBucket("StormItemTag")

	ctx = context.WithValue(ctx, "store", store)

	index, _ := CreateBleveIndex("wdid.test", true)
	ctx = context.WithValue(ctx, "index", index)

	f(ctx, store)
	db.Close()
	index.Close()

}

// mostRecentItem returns the most recent item chronologically in the store, up until time.Now()
// it does not deal with insertion order
func mostRecentItem(store Store) *Item {
	items, err := store.List(NewTimespan(time.Unix(0, 0), time.Now()))
	if err != nil {
		panic(err)
	}
	return items[len(items)-1]
}
