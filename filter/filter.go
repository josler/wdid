package filter

import (
	"github.com/asdine/storm/q"
)

type Filter interface {
	QueryItems() ([]q.Matcher, error)
}
