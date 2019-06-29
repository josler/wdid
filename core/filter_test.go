package core

import (
	"context"
	"testing"

	"github.com/josler/wdid/filter"
	"gotest.tools/assert"
)

func TestStatusFilterFunctionSplit(t *testing.T) {
	statusFilter, _ := StatusFilterFn(filter.FilterNe, "done|waiting")
	assert.Equal(t, statusFilter.(*StatusFilter).comparison, filter.FilterNe, "doesn't set comparison correctly")
	assert.DeepEqual(t, statusFilter.(*StatusFilter).statuses, []string{"done", "waiting"})
}

func TestStatusFilterFunctionError(t *testing.T) {
	_, err := StatusFilterFn(filter.FilterEq, "foobar")
	assert.Error(t, err, "invalid status")
}

func TestTagFilterFunction(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		tagFilter, _ := TagFilterFn(store)(filter.FilterEq, "#foo")
		assert.Equal(t, tagFilter.(*TagFilter).tagName, "#foo")
	})
}

func TestDateFilterFunction(t *testing.T) {
	dateFilter, _ := DateFilterFn(filter.FilterEq, "2019-05-18")
	timespan, _ := TimeParser{Input: "2019-05-18"}.Parse()
	assert.DeepEqual(t, dateFilter.(*DateFilter).timespan, timespan)
}

func TestDateFilterGtFunction(t *testing.T) {
	dateFilter, _ := DateFilterFn(filter.FilterGt, "2019-05-18")
	timespan, _ := TimeParser{Input: "2019-05-18"}.Parse()
	timespan.End = Timespan{}.LatestTime()
	assert.DeepEqual(t, dateFilter.(*DateFilter).timespan, timespan)
}

func TestDateFilterLtFunction(t *testing.T) {
	dateFilter, _ := DateFilterFn(filter.FilterLt, "2019-05-18")
	timespan, _ := TimeParser{Input: "2019-05-18"}.Parse()
	timespan.Start = Timespan{}.EarliestTime()
	assert.DeepEqual(t, dateFilter.(*DateFilter).timespan, timespan)
}

func TestDateFilterNeFunction(t *testing.T) {
	_, err := DateFilterFn(filter.FilterNe, "2019-05-18")
	assert.Error(t, err, "date filter does not support comparison 'ne'")
}
