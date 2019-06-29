package core

import (
	"errors"
	"strings"

	"github.com/josler/wdid/filter"
	"github.com/josler/wdid/parser"
)

type DateFilter struct {
	timespan *Timespan
}

func NewDateFilter(comparison filter.FilterComparison, timespan *Timespan) *DateFilter {
	return &DateFilter{timespan: timespan}
}

func DateFilterFn(comparison filter.FilterComparison, val string) (filter.Filter, error) {
	from, err := TimeParser{Input: val}.Parse()
	if err != nil {
		return nil, err
	}
	switch comparison {
	case filter.FilterGt:
		from.End = Timespan{}.LatestTime()
	case filter.FilterLt:
		from.Start = Timespan{}.EarliestTime()
	case filter.FilterNe:
		return nil, errors.New("date filter does not support comparison 'ne'")
	}

	return NewDateFilter(comparison, from), nil
}

func (dateFilter *DateFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)
	return (stormItem.Datetime >= dateFilter.timespan.Start.Unix() && stormItem.Datetime <= dateFilter.timespan.End.Unix()), nil
}

type StatusFilter struct {
	comparison filter.FilterComparison
	statuses   []string
}

func NewStatusFilter(comparison filter.FilterComparison, statuses ...string) *StatusFilter {
	return &StatusFilter{comparison: comparison, statuses: statuses}
}

func StatusFilterFn(comparison filter.FilterComparison, val string) (filter.Filter, error) {
	switch comparison {
	case filter.FilterGt, filter.FilterLt:
		return nil, errors.New("status filter does not support comparison > or <")
	}

	validStatuses := map[string]struct{}{WaitingStatus: struct{}{}, SkippedStatus: struct{}{}, DoneStatus: struct{}{}, BumpedStatus: struct{}{}}
	// allow usage of OR split - beta feature
	statusValues := strings.Split(val, "|")
	for _, val := range statusValues {
		if _, ok := validStatuses[val]; !ok {
			return nil, errors.New("invalid status")
		}
	}
	return NewStatusFilter(comparison, statusValues...), nil
}

func (statusFilter *StatusFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)

	if statusFilter.comparison == filter.FilterEq {
		for _, okStatus := range statusFilter.statuses {
			// for an EQ comparison, always return true if any candidate statuses match
			// the status of this item
			if stormItem.Status == okStatus {
				return true, nil
			}
		}
		return false, nil
	}

	if statusFilter.comparison == filter.FilterNe {
		for _, okStatus := range statusFilter.statuses {
			// for an NE comparison, always return false if any candidate statuses match
			// the status of this item
			if stormItem.Status == okStatus {
				return false, nil
			}
		}
		return true, nil
	}

	return false, errors.New("unrecognized comparison")
}

type TagFilter struct {
	store      Store
	comparison filter.FilterComparison
	tagName    string
}

func NewTagFilter(store Store, comparison filter.FilterComparison, name string) *TagFilter {
	return &TagFilter{store: store, comparison: comparison, tagName: name}
}

func TagFilterFn(store Store) parser.ToFilterFn {
	return func(comparison filter.FilterComparison, val string) (filter.Filter, error) {
		switch comparison {
		case filter.FilterGt, filter.FilterLt:
			return nil, errors.New("tag filter does not support > or <")
		}
		return NewTagFilter(store, comparison, val), nil
	}
}

func (tagFilter *TagFilter) Match(i interface{}) (bool, error) {
	stormItem := i.(StormItem)
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(stormItem.Data)
	if err != nil {
		return false, err
	}

	if tagFilter.comparison == filter.FilterEq {
		for _, res := range tokenResult.Tags {
			// more matching eq, if we ever do find a match
			// evaluate to true
			if tagFilter.tagName == res {
				return true, nil
			}
		}
		return false, nil
	}

	if tagFilter.comparison == filter.FilterNe {
		for _, res := range tokenResult.Tags {
			// for matching the negative, if we ever _do_ find a match,
			// it should evaluate to false
			if tagFilter.tagName == res {
				return false, nil
			}
		}
		// na matches is true
		return true, nil
	}

	return false, errors.New("unrecognized comparison")
}
