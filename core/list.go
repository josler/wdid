package core

import (
	"context"

	"github.com/josler/wdid/parser"
)

func List(ctx context.Context, timeString string, filterString string, statuses ...string) error {
	store := ctx.Value("store").(Store)
	itemPrinter := NewItemPrinter(ctx)

	var items []*Item
	var err error

	if filterString != "" {
		items, err = listFromFilters(store, filterString)
	} else {
		items, err = listFromFlags(store, timeString, statuses...)
	}

	itemPrinter.Print(items...)
	return err
}

func listFromFilters(store Store, filterString string) ([]*Item, error) {
	bs := store.(*BoltStore)

	p := &parser.Parser{}
	p.RegisterToFilter("tag", TagFilterFn(bs))
	p.RegisterToFilter("status", StatusFilterFn)
	p.RegisterToFilter("time", DateFilterFn)

	filters, err := p.Parse(filterString)
	if err != nil {
		return []*Item{}, err
	}

	return bs.ListFilters(filters)
}

func listFromFlags(store Store, timeString string, statuses ...string) ([]*Item, error) {
	from, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return []*Item{}, err
	}
	return store.List(from, statuses...)
}
