package core

import (
	"strings"
	"testing"
	"time"
)

func TestBase36(t *testing.T) {
	testBase36Matches(t, 0, '0')
	testBase36Matches(t, 1, '1')
	testBase36Matches(t, 12, 'c')
	testBase36Matches(t, 35, 'z')
	testBase36Matches(t, 36, '0')
	testBase36Matches(t, 71, 'z')
}

func TestGenerateIDFromTime(t *testing.T) {
	expectedSuffix := "i3m"
	ti := timeAt("2018-03-22 00:00:00 -0400 EDT")
	output := GenerateID(ti)

	if !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("With time '%v', expected '%s' to end with '%s'", ti, output, expectedSuffix)
	}
}

func TestNewItemHasTimeBasedID(t *testing.T) {
	expectedSuffix := "i3m"
	ti := timeAt("2018-03-22 00:00:00 -0400 EDT")
	item := NewItem("foobar", ti)
	if !strings.HasSuffix(item.ID(), expectedSuffix) {
		t.Errorf("With time '%v', expected '%s' to end with '%s'", ti, item.Time(), expectedSuffix)
	}
}

func TestBump(t *testing.T) {
	bumpedItem := NewItem("foobar", time.Now())
	newItem := bumpedItem.Bump(time.Now())
	if newItem.PreviousID() != bumpedItem.ID() {
		t.Errorf("Bumped item ID %s and new item PreviousID %s do not match", bumpedItem.ID(), newItem.PreviousID())
	}
	if newItem.ID() != bumpedItem.NextID() {
		t.Errorf("Bumped item NextID %s and new item ID %s do not match", bumpedItem.NextID(), newItem.ID())
	}
	if bumpedItem.Status() != BumpedStatus {
		t.Errorf("Bumped item status was %s not bumped", bumpedItem.Status())
	}
}

func TestDo(t *testing.T) {
	item := NewItem("foobar", time.Now())
	item.Do()
	if item.Status() != DoneStatus {
		t.Errorf("item was not marked done")
	}
}

func TestDoBumped(t *testing.T) {
	bumpedItem := NewItem("foobar", time.Now())
	bumpedItem.Bump(time.Now())
	bumpedItem.Do()
	if bumpedItem.Status() != BumpedStatus {
		t.Errorf("Bumped item could be marked as done")
	}
}

func TestSkip(t *testing.T) {
	item := NewItem("foobar", time.Now())
	item.Skip()
	if item.Status() != SkippedStatus {
		t.Errorf("item was not marked skipped")
	}
}

func TestSkipBumped(t *testing.T) {
	bumpedItem := NewItem("foobar", time.Now())
	bumpedItem.Bump(time.Now())
	bumpedItem.Skip()
	if bumpedItem.Status() != BumpedStatus {
		t.Errorf("Bumped item could be marked as skipped")
	}
}

func testBase36Matches(t *testing.T, input int, expected rune) {
	output := Base36(input)
	if output != expected {
		t.Errorf("With input '%d', expected %c, got %c", input, expected, output)
	}
}
