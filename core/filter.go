package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/josler/wdid/filter"
	"github.com/josler/wdid/parser"
)

func DefaultParser(store Store) *parser.Parser {
	p := &parser.Parser{}
	p.RegisterToFilter("tag", TagFilterFn(store))
	p.RegisterToFilter("status", StatusFilterFn)
	p.RegisterToFilter("time", DateFilterFn)
	p.RegisterToFilter("group", GroupFilterFn(store))
	p.RegisterToFilter("kind", KindFilterFn)
	return p
}

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

func (dateFilter *DateFilter) Match(matchable filter.Matchable) (bool, error) {
	return (matchable.Datetime() >= dateFilter.timespan.Start.Unix() && matchable.Datetime() <= dateFilter.timespan.End.Unix()), nil
}

func (dateFilter *DateFilter) String() string {
	return fmt.Sprintf("Between %v and %v", dateFilter.timespan.Start, dateFilter.timespan.End)
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

	validStatuses := map[string]struct{}{WaitingStatus: {}, SkippedStatus: {}, DoneStatus: {}, BumpedStatus: {}}
	// allow usage of OR split - beta feature
	statusValues := strings.Split(val, "|")
	for _, val := range statusValues {
		if _, ok := validStatuses[val]; !ok {
			return nil, errors.New("invalid status")
		}
	}
	return NewStatusFilter(comparison, statusValues...), nil
}

func (statusFilter *StatusFilter) Match(matchable filter.Matchable) (bool, error) {
	if statusFilter.comparison == filter.FilterEq {
		for _, okStatus := range statusFilter.statuses {
			// for an EQ comparison, always return true if any candidate statuses match
			// the status of this item
			if matchable.Status() == okStatus {
				return true, nil
			}
		}
		return false, nil
	}

	if statusFilter.comparison == filter.FilterNe {
		for _, okStatus := range statusFilter.statuses {
			// for an NE comparison, always return false if any candidate statuses match
			// the status of this item
			if matchable.Status() == okStatus {
				return false, nil
			}
		}
		return true, nil
	}

	return false, errors.New("unrecognized comparison")
}

func (statusFilter *StatusFilter) String() string {
	return fmt.Sprintf("Status %v %v", statusFilter.comparison, statusFilter.statuses)
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

func (tagFilter *TagFilter) Match(matchable filter.Matchable) (bool, error) {
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(matchable.Data())
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

func (tagFilter *TagFilter) String() string {
	return fmt.Sprintf("Tag %v %s", tagFilter.comparison, tagFilter.tagName)
}

type GroupFilter struct {
	comparison   filter.FilterComparison
	name         string
	groupFilters []filter.Filter
}

func NewGroupFilter(comparison filter.FilterComparison, name string, filters []filter.Filter) *GroupFilter {
	return &GroupFilter{comparison: comparison, name: name, groupFilters: filters}
}

func GroupFilterFn(store Store) parser.ToFilterFn {
	return func(comparison filter.FilterComparison, val string) (filter.Filter, error) {
		switch comparison {
		case filter.FilterGt, filter.FilterLt:
			return nil, errors.New("group filter does not support > or <")
		}

		group, err := store.FindGroupByName(val)
		if err != nil {
			return nil, fmt.Errorf("Failed to find group by name: %w", err)
		}

		filters, err := group.Filters(store)
		if err != nil {
			return nil, err
		}

		return &GroupFilter{comparison: comparison, name: group.Name, groupFilters: filters}, nil
	}
}

func (groupFilter *GroupFilter) Match(matchable filter.Matchable) (bool, error) {

	if groupFilter.comparison == filter.FilterEq {
		for _, filter := range groupFilter.groupFilters {
			innerMatch, err := filter.Match(matchable)
			// if doesn't match an inner filter or if error
			// then we can't match EQ
			if !innerMatch || err != nil {
				return false, err
			}
		}
		// has matched all
		return true, nil
	}
	if groupFilter.comparison == filter.FilterNe {
		for _, filter := range groupFilter.groupFilters {
			innerMatch, err := filter.Match(matchable)
			if innerMatch || err != nil { // if matches inner or if error
				// then we know we can't fully match NE
				return false, err
			}
		}
		// has matched none
		return true, nil
	}
	return false, errors.New("unrecognized comparison")
}

func (groupFilter *GroupFilter) String() string {
	return fmt.Sprintf("Group %v %s", groupFilter.comparison, groupFilter.name)
}

type KindFilter struct {
	comparison filter.FilterComparison
	matchKind  Kind
}

func NewKindFilter(comparison filter.FilterComparison, matchKind Kind) *KindFilter {
	return &KindFilter{
		comparison: comparison,
		matchKind:  matchKind,
	}
}

func KindFilterFn(comparison filter.FilterComparison, matchKind string) (filter.Filter, error) {
	switch comparison {
	case filter.FilterGt, filter.FilterLt:
		return nil, errors.New("kind filter does not support > or <")
	}
	kind := StringToKind(matchKind)
	if kind <= 0 {
		return nil, fmt.Errorf("kind %q not found", matchKind)
	}
	return NewKindFilter(comparison, StringToKind(matchKind)), nil
}

func (kindFilter *KindFilter) Match(matchable filter.Matchable) (bool, error) {
	kind := Kind(matchable.Kind())
	if kind <= 0 {
		kind = Task
	}

	switch kindFilter.comparison {
	case filter.FilterEq:
		return (kind == kindFilter.matchKind), nil
	case filter.FilterNe:
		return (kind != kindFilter.matchKind), nil
	}
	return false, fmt.Errorf("failed to compare kind correctly")
}
