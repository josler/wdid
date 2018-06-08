package core

import (
	"testing"
	"time"
)

func TestParseInt(t *testing.T) {
	ref := timeAt("2018-03-23 17:53:30 -0400 EDT")
	testInputOutput(t, ref, "0", timeAt("2018-03-23 00:00:00 -0400 EDT"), timeAt("2018-03-23 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "1", timeAt("2018-03-22 00:00:00 -0400 EDT"), timeAt("2018-03-23 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "6", timeAt("2018-03-17 00:00:00 -0400 EDT"), timeAt("2018-03-23 23:59:59 -0400 EDT"))
}

func TestParseWord(t *testing.T) {
	ref := timeAt("2018-03-23 17:53:30 -0400 EDT")
	testInputOutput(t, ref, "now", ref, ref)
	testInputOutput(t, ref, "day", timeAt("2018-03-23 00:00:00 -0400 EDT"), timeAt("2018-03-23 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "week", timeAt("2018-03-19 00:00:00 -0400 EDT"), timeAt("2018-03-25 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "month", timeAt("2018-03-01 00:00:00 -0500 EDT"), timeAt("2018-03-31 23:59:59 -0400 EDT"))

	testInputOutput(t, ref, "tomorrow", timeAt("2018-03-24 00:00:00 -0400 EDT"), timeAt("2018-03-24 23:59:59 -0400 EDT"))
}

func TestParseFormat(t *testing.T) {
	ref := timeAt("2018-03-23 17:53:30 -0400 EDT")
	testInputOutput(t, ref, "2018-03-23", timeAt("2018-03-23 00:00:00 -0400 EDT"), timeAt("2018-03-23 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "2018-03-22", timeAt("2018-03-22 00:00:00 -0400 EDT"), timeAt("2018-03-22 23:59:59 -0400 EDT"))
	testInputOutput(t, ref, "2018-02-01", timeAt("2018-02-01 00:00:00 -0400 EDT"), timeAt("2018-02-01 23:59:59 -0400 EDT"))

	testInputOutput(t, ref, "2018-02-01T16:20", timeAt("2018-02-01 16:20:00 -0400 EDT"), timeAt("2018-02-01 23:59:59 -0400 EDT"))
}

func timeAt(rfc string) time.Time {
	ret, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", rfc)
	return ret
}

func testInputOutput(t *testing.T, referenceTime time.Time, input string, expectedStart, expectedEnd time.Time) {
	tp := TimeParser{Input: input, startTime: referenceTime}
	output, err := tp.Parse()
	if err != nil {
		t.Errorf("Input '%s' failed with error %v", input, err)
	}
	if !output.Start.Equal(expectedStart) {
		t.Errorf("Input start '%s' failed to match expected %v, was %v", input, expectedStart, output.Start)
	}
	if !output.End.Equal(expectedEnd) {
		t.Errorf("Input end '%s' failed to match expected %v, was %v", input, expectedEnd, output.End)
	}
}
