package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/josler/wdid/filter"
)

func List(ctx context.Context, argString string, groupString string) error {
	v := ctx.Value("verbose")
	isVerbose := v != nil && v.(bool)

	store := ctx.Value("store").(Store)
	itemPrinter := NewItemPrinter(ctx)

	var items []*Item
	var err error

	if groupString != "" {
		group, err := store.FindGroupByName(groupString)
		if err != nil {
			return err
		}

		argString = strings.Join([]string{argString, group.FilterString}, ",")
		argString = strings.TrimPrefix(argString, ",")
	}

	items, err = listFromFilters(store, argString, isVerbose)
	if err != nil {
		items, err = listFromTimeString(store, argString)
	}

	itemPrinter.Print(items...)
	return err
}

func listFromFilters(store Store, filterString string, isVerbose bool) ([]*Item, error) {
	p := DefaultParser(store)
	filters, err := p.Parse(filterString)
	if err != nil {
		return []*Item{}, err
	}

	if isVerbose {
		fmt.Println("Filters:")
		for _, filter := range filters {
			fmt.Println(filter)
		}
		fmt.Println("")
	}

	return store.ListFilters(filters)
}

func listFromTimeString(store Store, timeString string) ([]*Item, error) {
	from, err := TimeParser{Input: timeString}.Parse()
	if err != nil {
		return []*Item{}, err
	}

	filters := []filter.Filter{NewDateFilter(filter.FilterEq, from)}
	return store.ListFilters(filters)
}
