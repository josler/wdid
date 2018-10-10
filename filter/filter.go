package filter

import (
	"github.com/asdine/storm/q"
	"github.com/blevesearch/bleve/search/query"
)

type Filter interface {
	QueryItems() ([]q.Matcher, error)
	BleveQuery() (query.Query, error)
}
