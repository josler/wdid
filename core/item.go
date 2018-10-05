package core

import (
	"fmt"
	"time"

	"gitlab.com/josler/wdid/parser"
)

const (
	WaitingStatus = "waiting"
	DoneStatus    = "done"
	SkippedStatus = "skipped"
	BumpedStatus  = "bumped"
)

type Item struct {
	internalID string
	id         string
	nextID     string // ref to next if bumped
	previousID string // ref to previous if bumped
	data       string
	tags       []*Tag
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

func (i *Item) Tags() []*Tag {
	if i.tags == nil {
		i.generateMetadata()
	}
	return i.tags
}

func (i *Item) Status() string {
	return i.status
}

func (i *Item) Time() time.Time {
	return i.datetime
}

func (i *Item) Do() {
	i.status = DoneStatus
}

func (i *Item) Skip() {
	i.status = SkippedStatus
}

func (i *Item) Bump(newTime time.Time) *Item {
	i.status = BumpedStatus
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

func (i *Item) generateMetadata() {
	i.tags = []*Tag{} // always init
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(i.data)
	if err == nil {
		for _, resultTag := range tokenResult.Tags {
			i.tags = append(i.tags, NewTag(resultTag))
		}
	}
}

func NewItem(data string, at time.Time) *Item {
	return &Item{id: GenerateID(at), data: data, status: WaitingStatus, datetime: at}
}
