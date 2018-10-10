package core

import (
	"errors"

	"github.com/asdine/storm/q"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
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

func (dateFilter *DateFilter) BleveQuery() (query.Query, error) {
	start := float64(dateFilter.timespan.Start.Unix())
	end := float64(dateFilter.timespan.End.Unix())
	inc := true
	dateQuery := bleve.NewNumericRangeInclusiveQuery(&start, &end, &inc, &inc)
	dateQuery.SetField("Datetime")
	return dateQuery, nil
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

func (statusFilter *StatusFilter) BleveQuery() (query.Query, error) {
	statusQuery := bleve.NewMatchQuery(statusFilter.statuses[0])
	statusQuery.SetField("Status")
	return statusQuery, nil
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
	ids := []string{}

	for _, item := range items {
		ids = append(ids, item.ID())
	}
	if len(ids) == 0 {
		return []q.Matcher{}, errors.New("failed to find items with tag")
	}
	return []q.Matcher{
		q.In("ID", ids),
	}, nil
}

func (tagFilter *TagFilter) BleveQuery() (query.Query, error) {
	found, err := tagFilter.store.FindTag(tagFilter.tagName)
	if err != nil {
		return nil, err
	}
	tagQuery := bleve.NewMatchQuery(found.internalID)
	//tagQuery.SetField("Tag.TagID")
	tagQuery.SetField("TagIDs")
	return tagQuery, nil
}
