package core

import (
	"errors"

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

func (dateFilter *DateFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)
	return (stormItem.Datetime >= dateFilter.timespan.Start.Unix() && stormItem.Datetime <= dateFilter.timespan.End.Unix()), nil
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

func (statusFilter *StatusFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)
	for _, okStatus := range statusFilter.statuses {
		if stormItem.Status == okStatus {
			return true, nil
		}
	}
	return false, nil
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

func (tagFilter *TagFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(stormItem.Data)
	if err != nil {
		return false, err
	}

	for _, res := range tokenResult.Tags {
		if res == tagFilter.tagName {
			return true, nil
		}
	}
	return false, nil
}
