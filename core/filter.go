package core

import (
	"errors"

	"github.com/asdine/storm/q"
	"gitlab.com/josler/wdid/filter"
	"gitlab.com/josler/wdid/parser"
)

type DateFilter struct {
	timespan *Timespan
}

func NewDateFilter(timespan *Timespan) *DateFilter {
	return &DateFilter{timespan: timespan}
}

func DateFilterFn(val string) (filter.Filter, error) {
	from, err := TimeParser{Input: val}.Parse()
	if err != nil {
		return nil, err
	}
	return NewDateFilter(from), nil
}

func (dateFilter *DateFilter) QueryItems() ([]q.Matcher, error) {
	return []q.Matcher{
		q.Gte("Datetime", dateFilter.timespan.Start.Unix()),
		q.Lte("Datetime", dateFilter.timespan.End.Unix()),
	}, nil
}

type StatusFilter struct {
	statuses []string
}

func NewStatusFilter(statuses ...string) *StatusFilter {
	return &StatusFilter{statuses: statuses}
}

func StatusFilterFn(val string) (filter.Filter, error) {
	validStatuses := map[string]struct{}{WaitingStatus: struct{}{}, SkippedStatus: struct{}{}, DoneStatus: struct{}{}, BumpedStatus: struct{}{}}
	if _, ok := validStatuses[val]; !ok {
		return nil, errors.New("invalid status")
	}
	return NewStatusFilter(val), nil
}

func (statusFilter *StatusFilter) QueryItems() ([]q.Matcher, error) {
	return []q.Matcher{
		q.In("Status", statusFilter.statuses),
	}, nil
}

type TagFilter struct {
	store   Store
	tagName string
}

func NewTagFilter(store Store, name string) *TagFilter {
	return &TagFilter{store: store, tagName: name}
}

func TagFilterFn(store Store) parser.ToFilterFn {
	return func(val string) (filter.Filter, error) {
		return NewTagFilter(store, val), nil
	}
}

type MapMatcher struct {
	elements map[string]bool
}

func (mm *MapMatcher) Add(element string) {
	if mm.elements == nil {
		mm.elements = map[string]bool{}
	}
	mm.elements[element] = true
}

func (mm *MapMatcher) Match(i interface{}) (bool, error) {
	storeItem := i.(StormItem)
	_, ok := mm.elements[storeItem.ID]

	return ok, nil
}

func (tagFilter *TagFilter) QueryItems() ([]q.Matcher, error) {
	// find tag id
	tag, err := tagFilter.store.FindTag(tagFilter.tagName)
	if err != nil {
		return []q.Matcher{}, errors.New("failed to find tag")
	}
	items, err := tagFilter.store.FindItemsWithTag(tag, -1)
	if err != nil {
		return []q.Matcher{}, errors.New("failed to find items with tag")
	}
	if len(items) == 0 {
		return []q.Matcher{}, errors.New("failed to find items with tag")
	}
	mm := &MapMatcher{}

	//ids := []string{}
	for _, item := range items {
		//ids = append(ids, item.ID())
		mm.Add(item.ID())
	}
	return []q.Matcher{mm}, nil
	//return []q.Matcher{
	//q.In("ID", ids),
	//}, nil
}
