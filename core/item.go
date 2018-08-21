package core

import (
	"fmt"
	"math/rand"
	"time"
)

type Item struct {
	internalID string
	id         string
	nextID     string // ref to next if bumped
	previousID string // ref to previous if bumped
	data       string
	status     string
	datetime   time.Time
}

func (i *Item) ID() string {
	return i.id
}

func (i *Item) NextID() string {
	return i.nextID
}

func (i *Item) PreviousID() string {
	return i.previousID
}

func (i *Item) Data() string {
	return i.data
}

func (i *Item) Status() string {
	return i.status
}

func (i *Item) Time() time.Time {
	return i.datetime
}

func (i *Item) Do() {
	i.status = "done"
}

func (i *Item) Skip() {
	i.status = "skipped"
}

func (i *Item) Bump(newTime time.Time) *Item {
	i.status = "bumped"
	newItem := NewItem(i.data, newTime)
	i.nextID = newItem.ID()
	newItem.previousID = i.ID()
	return newItem
}

func (i *Item) String() string {
	return fmt.Sprintf("%s: %s (%s) %v", i.ID(), i.Data(), i.Status(), i.Time())
}

func (i *Item) ResetInternalID() {
	i.internalID = ""
}

// manually set the ID, useful for testing!
func (i *Item) SetID(id string) {
	i.id = id[:MAX_ID_LENGTH]
}

func NewItem(data string, at time.Time) *Item {
	return &Item{id: GenerateID(at), data: data, status: "waiting", datetime: at}
}

// GenerateID generates a 6 digit ID - with the last three digits sortable by Year, Month, Day
// Format: RRRYMD - where R is a random base36 rune.
// This means, that in any given day, there's 36^3 chance of a random collision - acceptable for this.
func GenerateID(at time.Time) string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return fmt.Sprintf("%c%c%c%s", Base36(r1.Intn(35)), Base36(r1.Intn(35)), Base36(r1.Intn(35)), IDSuffixForDate(at))
}

// numbers above 35 translate to the value of remainder. i.e. 36 is 0, 71 is z, 72 is 0...
func Base36(in int) rune {
	charMap := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	return charMap[(in % 36)]
}

func IDSuffixForDate(t time.Time) string {
	return fmt.Sprintf("%c%c%c", Base36(int(t.Year()-2000)), Base36(int(t.Month())), Base36(int(t.Day())))
}
