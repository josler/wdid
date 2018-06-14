package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	case "today":
		return NewTimespan(tp.startOfDay(tp.startTime), tp.endOfDay(tp.startTime)), nil
	case "tomorrow":
		startOfTomorrow := tp.startOfDay(tp.startTime).AddDate(0, 0, 1)
		return NewTimespan(startOfTomorrow, tp.endOfDay(startOfTomorrow)), nil
	case "yesterday":
		startOfYesterday := tp.startOfDay(tp.startTime).Add(-24 * time.Hour)
		return NewTimespan(startOfYesterday, tp.endOfDay(startOfYesterday)), nil
	}

	weekday, err := tp.getWeekday(tp.Input)
	if err == nil {
		return tp.nextOccuranceOfWeekday(tp.startTime, weekday, 24), nil
	}

	// try to parse a weekday phrase
	splitStrings := strings.Split(tp.Input, " ")
	if len(splitStrings) == 2 {
		weekday, err := tp.getWeekday(splitStrings[1])
		if err == nil {
			// parse "<offset> <weekday>"
			switch splitStrings[0] {
			case "last":
				return tp.nextOccuranceOfWeekday(tp.startOfWeek(tp.startTime).AddDate(0, 0, -1), weekday, -24), nil
			case "this":
				return tp.nextOccuranceOfWeekday(tp.startOfWeek(tp.startTime), weekday, 24), nil
			case "next":
				return tp.nextOccuranceOfWeekday(tp.endOfWeek(tp.startTime).AddDate(0, 0, 1), weekday, 24), nil
			}
		}
		switch splitStrings[1] {
		// parse "<offset> <duration>"
		case "week":
			switch splitStrings[0] {
			case "last":
				lastWeek := tp.startTime.AddDate(0, 0, -7)
				return NewTimespan(tp.startOfWeek(lastWeek), tp.endOfWeek(lastWeek)), nil
			case "this":
				return NewTimespan(tp.startOfWeek(tp.startTime), tp.endOfWeek(tp.startTime)), nil
			case "next":
				nextWeek := tp.startTime.AddDate(0, 0, 7)
				return NewTimespan(tp.startOfWeek(nextWeek), tp.endOfWeek(nextWeek)), nil
			}
		case "month":
			switch splitStrings[0] {
			case "last":
				lastMonth := tp.startTime.AddDate(0, -1, 0)
				return NewTimespan(tp.startOfMonth(lastMonth), tp.endOfMonth(lastMonth)), nil
			case "this":
				return NewTimespan(tp.startOfMonth(tp.startTime), tp.endOfMonth(tp.startTime)), nil
			case "next":
				nextMonth := tp.startTime.AddDate(0, 1, 0)
				return NewTimespan(tp.startOfMonth(nextMonth), tp.endOfMonth(nextMonth)), nil
			}
		default:
			return NewTimespan(tp.startTime, tp.startTime), errors.New(fmt.Sprintf("failed to parse time with input: %s", tp.Input))
		}

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

func (tp TimeParser) nextOccuranceOfWeekday(startAt time.Time, weekday time.Weekday, jump time.Duration) *Timespan {
	startTime := startAt
	for {
		if startTime.Weekday() == weekday {
			return NewTimespan(tp.startOfDay(startTime), tp.endOfDay(startTime))
		}
		startTime = startTime.Add(jump * time.Hour)
	}
}

func (tp TimeParser) getWeekday(word string) (time.Weekday, error) {
	weekdays := map[string]time.Weekday{
		"mon":       time.Monday,
		"monday":    time.Monday,
		"tue":       time.Tuesday,
		"tues":      time.Tuesday,
		"tuesday":   time.Tuesday,
		"wed":       time.Wednesday,
		"weds":      time.Wednesday,
		"wednesday": time.Wednesday,
		"thu":       time.Thursday,
		"thur":      time.Thursday,
		"thurs":     time.Thursday,
		"thursday":  time.Thursday,
		"fri":       time.Friday,
		"friday":    time.Friday,
		"sat":       time.Saturday,
		"saturday":  time.Saturday,
		"sun":       time.Sunday,
		"sunday":    time.Sunday,
	}
	found, ok := weekdays[word]
	if !ok {
		return time.Monday, errors.New("weekday not found")
	}
	return found, nil
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
