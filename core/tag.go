package core

import (
	"strings"
	"time"
)

type Tag struct {
	internalID string
	name       string
	createdAt  time.Time
}

func NewTag(name string) *Tag {
	return &Tag{name: name, createdAt: time.Now()}
}

func (t *Tag) Name() string {
	return t.name
}

func (t *Tag) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Tag) TagType() string {
	if strings.HasPrefix(t.Name(), "#") {
		return "hashtag"
	} else if strings.HasPrefix(t.Name(), "@") {
		return "mention"
	} else {
		return "plain"
	}
}

func (t *Tag) String() string {
	return t.Name()
}
