package core

import (
	"fmt"
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

func (t *Tag) Tag() string {
	return fmt.Sprintf("%s: %v", t.Name(), t.CreatedAt())
}
