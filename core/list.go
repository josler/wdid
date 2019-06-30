package core

import (
	"context"
	"strings"

	"github.com/josler/wdid/parser"
)

func List(ctx context.Context, timeString string, filterString string, groupString string, statuses ...string) error {
	store := ctx.Value("store").(Store)
	itemPrinter := NewItemPrinter(ctx)

	var items []*Item
	var err error

	if groupString != "" {
		group, err := store.FindGroupByName(groupString)
		if err != nil {
			return err
		}

		filterString = strings.Join([]string{filterString, group.FilterString}, ",")
		filterString = strings.TrimPrefix(filterString, ",")
	}

	if filterString != "" {
		items, err = listFromFilters(store, filterString)
	} else {
		items, err = listFromFlags(store, timeString, statuses...)
	}

	itemPrinter.Print(items...)
	return err
}

func listFromFilters(store Store, filterString string) ([]*Item, error) {
	p := &parser.Parser{}
	p.RegisterToFilter("tag", TagFilterFn(store))
	p.RegisterToFilter("status", StatusFilterFn)
	p.RegisterToFilter("time", DateFilterFn)

	filters, err := p.Parse(filterString)
	if err != nil {
		return []*Item{}, err
	}

	return store.ListFilters(filters)
}

func listFromFlags(store Store, timeString string, statuses ...string) ([]*Item, error) {
	from, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return []*Item{}, err
	}
	return store.List(from, statuses...)
}
