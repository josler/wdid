package core

import (
	"fmt"
	"sort"
	"time"

	"github.com/josler/wdid/parser"
)

const (
	NoStatus      = "none"
	WaitingStatus = "waiting"
	DoneStatus    = "done"
	SkippedStatus = "skipped"
	BumpedStatus  = "bumped"
)

type Kind int

const (
	_ Kind = iota
	Task
	Note
)

func (k Kind) String() string {
	switch k {
	case Task:
		return "task"
	case Note:
		return "note"
	}
	return ""
}

func StringToKind(kindstring string) Kind {
	switch kindstring {
	case "task":
		return Task
	case "note":
		return Note
	}
	return 0
}

type Item struct {
	internalID  string
	id          string
	nextID      string // ref to next if bumped
	previousID  string // ref to previous if bumped
	data        string
	tags        []*Tag
	connections []string // connections to other items
	status      string
	datetime    time.Time
	kind        Kind
}

func (i *Item) ID() string {
	return i.id
}

func (i *Item) Kind() Kind {
	if i.kind == 0 {
		return Task
	}
	return i.kind
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
	if i.emptyMetadata() {
		i.generateMetadata()
	}
	sort.Slice(i.tags, func(j, k int) bool {
		return i.tags[j].Name() <= i.tags[k].Name()
	})
	return i.tags
}

func (i *Item) Connections() []string {
	if i.emptyMetadata() {
		i.generateMetadata()
	}
	return i.connections
}

func (i *Item) Status() string {
	return i.status
}

func (i *Item) Time() time.Time {
	return i.datetime
}

func (i *Item) Do() {
	if i.Kind() != Task {
		return // do does nothing with non tasks
	}
	if i.status != BumpedStatus {
		i.status = DoneStatus
	}
}

func (i *Item) Skip() {
	if i.Kind() != Task {
		return // do does nothing with non tasks
	}
	if i.status != BumpedStatus {
		i.status = SkippedStatus
	}
}

func (i *Item) Bump(newTime time.Time) *Item {
	if i.Kind() != Task {
		return i // do does nothing with non tasks
	}
	i.status = BumpedStatus
	newItem := NewTask(i.data, newTime)
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
	i.id = id[:MaxIDLength]
}

func (i *Item) SetKind(kind Kind) {
	i.kind = kind
	if i.kind == Note {
		i.status = NoStatus
	}
}

func (i *Item) generateMetadata() {
	i.tags = []*Tag{} // always init
	i.connections = []string{}
	tokenizer := &parser.Tokenizer{}
	tokenResult, err := tokenizer.Tokenize(i.data)
	if err == nil {
		for _, resultTag := range tokenResult.Tags {
			i.tags = append(i.tags, NewTag(resultTag))
		}

		if len(tokenResult.Connections) > 0 {
			i.connections = tokenResult.Connections
		}
	}
}

func (i *Item) emptyMetadata() bool {
	return i.tags == nil && i.connections == nil
}

func NewTask(data string, at time.Time) *Item {
	return &Item{id: GenerateID(at), data: data, status: WaitingStatus, datetime: at, kind: Task}
}

func NewNote(data string, at time.Time) *Item {
	return &Item{id: GenerateID(at), data: data, status: NoStatus, datetime: at, kind: Note}
}
