package core

import (
	"context"

	"github.com/blevesearch/bleve"
	"gitlab.com/josler/wdid/parser"
)

func List(ctx context.Context, timeString string, filterString string, statuses ...string) error {
	store := ctx.Value("store").(Store)
	index := ctx.Value("index").(bleve.Index)
	itemPrinter := NewItemPrinter(ctx)

	var items []*Item
	var err error

	if filterString != "" {
		items, err = listFromFilters(index, store, filterString)
	} else {
		items, err = listFromFlags(store, timeString, statuses...)
	}

	itemPrinter.Print(items...)
	index.Close()
	return err
}

func listFromFilters(index bleve.Index, store Store, filterString string) ([]*Item, error) {
	p := &parser.Parser{}
	p.RegisterToFilter("tag", TagFilterFn(store))
	p.RegisterToFilter("status", StatusFilterFn)
	p.RegisterToFilter("time", DateFilterFn)

	filters, err := p.Parse(filterString)
	if err != nil {
		return []*Item{}, err
	}
	return Query(index, filters...)
}

func listFromFlags(store Store, timeString string, statuses ...string) ([]*Item, error) {
	from, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return []*Item{}, err
	}
	return store.List(from, statuses...)
}
