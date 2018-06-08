package core

import "time"

type Timespan struct {
	Start time.Time
	End   time.Time
}

func NewTimespan(start, end time.Time) *Timespan {
	return &Timespan{Start: start, End: end}
}
