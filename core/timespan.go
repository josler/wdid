package core

import (
	"fmt"
	"time"
)

type Timespan struct {
	Start time.Time
	End   time.Time
}

func NewTimespan(start, end time.Time) *Timespan {
	return &Timespan{Start: start, End: end}
}

func (ts *Timespan) String() string {
	return fmt.Sprintf("Timespan { from: %s, to: %s }", ts.Start, ts.End)
}

func (ts Timespan) EarliestTime() time.Time {
	return time.Unix(0, 0)
}

func (ts Timespan) LatestTime() time.Time {
	// arbitrary future date
	t, _ := time.Parse("2006-01-02", "2076-07-12")
	return t
}
