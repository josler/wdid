package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type TimeParser struct {
	Input     string
	startTime time.Time
}

func (tp TimeParser) Parse() (*Timespan, error) {
	if tp.startTime.IsZero() {
		tp.startTime = time.Now()
	}

	// try to parse a relative int
	i, err := strconv.ParseInt(tp.Input, 10, 64)
	if err == nil {
		goBack := -24 * i
		timeAfterGoingBack := tp.startTime.Add(time.Duration(goBack) * time.Hour)
		return NewTimespan(tp.startOfDay(timeAfterGoingBack), tp.endOfDay(tp.startTime)), nil
	}

	// try to parse a word
	switch tp.Input {
	case "now":
		return NewTimespan(tp.startTime, tp.startTime), nil
	case "day":
		return NewTimespan(tp.startOfDay(tp.startTime), tp.endOfDay(tp.startTime)), nil
	case "week":
		return NewTimespan(tp.startOfWeek(tp.startTime), tp.endOfWeek(tp.startTime)), nil
	case "month":
		return NewTimespan(tp.startOfMonth(tp.startTime), tp.endOfMonth(tp.startTime)), nil
	case "tomorrow":
		startOfTomorrow := tp.startOfDay(tp.startTime).AddDate(0, 0, 1)
		return NewTimespan(startOfTomorrow, tp.endOfDay(startOfTomorrow)), nil
	}

	// try to parse a date from a formatted input
	// (manually append the relevant time zone)
	found, err := time.Parse("2006-01-02T15:04 -0700 MST", fmt.Sprintf("%s %s", tp.Input, tp.startTime.Format("-0700 MST")))
	if err == nil {
		return NewTimespan(found, tp.endOfDay(found)), nil
	}

	found, err = time.Parse("2006-01-02 -0700 MST", fmt.Sprintf("%s %s", tp.Input, tp.startTime.Format("-0700 MST")))
	if err == nil {
		return NewTimespan(tp.startOfDay(found), tp.endOfDay(found)), nil
	}

	return NewTimespan(tp.startTime, tp.startTime), errors.New(fmt.Sprintf("failed to parse time with input: %s", tp.Input))
}

func (tp TimeParser) startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (tp TimeParser) endOfDay(t time.Time) time.Time {
	return tp.startOfDay(t).AddDate(0, 0, 1).Add(-1 * time.Second)
}

func (tp TimeParser) startOfWeek(t time.Time) time.Time {
	dayOfWeek := t.Weekday()

	if dayOfWeek == 0 { // count sunday as last day, not first, because we're not *animals*
		dayOfWeek = 7
	}

	// go back dayOfWeek-1 days to find prev monday.
	return tp.startOfDay(t.Add(-24 * (time.Duration(dayOfWeek) - 1) * time.Hour))
}

func (tp TimeParser) endOfWeek(t time.Time) time.Time {
	return tp.startOfWeek(t).AddDate(0, 0, 7).Add(-1 * time.Second)
}

func (tp TimeParser) startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func (tp TimeParser) endOfMonth(t time.Time) time.Time {
	return tp.startOfMonth(t).AddDate(0, 1, 0).Add(-1 * time.Second)
}
