package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/blevesearch/bleve"
)

func TestListFromFiltersTag(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("another item @josler"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")
		index := ingestAll(store)

		filterString := "tag=#hashtag,time=now"
		items := getItemsFromFilters(t, index, store, filterString)
		if len(items) != 1 || items[0].Data() != "my item #hashtag" {
			t.Errorf("item not found")
		}
	})
}

func TestListFromFiltersStatus(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")
		item := mostRecentItem(store)
		Do(ctx, item.ID())
		index := ingestAll(store)

		filterString := "status=done"
		items := getItemsFromFilters(t, index, store, filterString)
		if len(items) != 1 || items[0].Data() != "my item #hashtag" {
			fmt.Println(items[0].Data())
			t.Errorf("item not found")
		}
	})
}

func TestListFromFiltersTime(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		time.Sleep(1 * time.Second)
		Add(ctx, strings.NewReader("another item @josler"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")
		index := ingestAll(store)

		filterString := "time=day"
		items := getItemsFromFilters(t, index, store, filterString)
		if len(items) != 2 {
			t.Fatalf("item not found")
		}
		if items[0].Data() != "my item #hashtag" {
			t.Errorf("wrong data")
		}
		if items[1].Data() != "another item @josler" {
			t.Errorf("wrong data")
		}
	})
}

func getItemsFromFilters(t *testing.T, index bleve.Index, store Store, filterString string) []*Item {
	var items []*Item
	items, err := listFromFilters(index, store, filterString)
	if err != nil {
		t.Errorf("error listing by filters")
	}
	return items
}

func ingestAll(store Store) bleve.Index {
	index, err := CreateBleveIndex("wdid.test", true)
	if err != nil {
		panic(err)
	}
	from, _ := TimeParser{Input: "14"}.Parse()
	items, _ := store.List(from)
	for _, item := range items {
		SaveBleve(index, store, item)
	}
	return index
}
