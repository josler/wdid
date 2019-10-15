package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestListFromFiltersTag(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("another item @josler"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")

		filterString := "tag=#hashtag,time=0"
		items := getItemsFromFilters(t, store, filterString)
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

		filterString := "status=done"
		items := getItemsFromFilters(t, store, filterString)
		if len(items) != 1 || items[0].Data() != "my item #hashtag" {
			fmt.Println(items[0].Data())
			t.Errorf("item not found")
		}
	})
}

func TestListFromFiltersWithSpaces(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("another item @josler"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")

		filterString := "time = 0"
		items := getItemsFromFilters(t, store, filterString)
		if len(items) != 2 {
			t.Errorf("item not found")
		}
		if items[0].Data() != "my item #hashtag" {
			t.Errorf("wrong data")
		}
		if items[1].Data() != "another item @josler" {
			t.Errorf("wrong data")
		}
	})
}

func TestListFromFiltersTime(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("another item @josler"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")

		filterString := "time=0"
		items := getItemsFromFilters(t, store, filterString)
		if len(items) != 2 {
			t.Errorf("item not found")
		}
		if items[0].Data() != "my item #hashtag" {
			t.Errorf("wrong data")
		}
		if items[1].Data() != "another item @josler" {
			t.Errorf("wrong data")
		}
	})
}

func TestListFromGroup(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my item #hashtag"), "now")
		Add(ctx, strings.NewReader("same #hashtag"), "2018-08-10")

		CreateGroup(ctx, "my group", "tag=#hashtag")
		err := List(ctx, "", "", "my group")
		if err != nil {
			t.Errorf("failed to list from group")
		}
	})
}

func getItemsFromFilters(t *testing.T, store Store, filterString string) []*Item {
	var items []*Item
	items, err := listFromFilters(store, filterString, false)
	if err != nil {
		t.Fatalf("error listing by filters")
	}
	return items
}
