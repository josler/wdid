package core

import (
	"context"
	"fmt"
	"time"

	"github.com/josler/wdid/filter"
)

type Group struct {
	internalID   string
	Name         string
	FilterString string
	CreatedAt    time.Time
}

func (g *Group) String() string {
	return fmt.Sprintf("%s: \"%s\"", g.Name, g.FilterString)
}

func NewGroup(name string, filterString string) *Group {
	return &Group{Name: name, FilterString: filterString, CreatedAt: time.Now()}
}

func (g *Group) Filters(store Store) ([]filter.Filter, error) {
	p := DefaultParser(store)
	filters, err := p.Parse(g.FilterString)
	if err != nil {
		return []filter.Filter{}, err
	}
	return filters, nil
}

func CreateGroup(ctx context.Context, name string, filterString string) error {
	store := ctx.Value("store").(Store)
	group := NewGroup(name, filterString)

	// validate filters
	_, err := group.Filters(store)
	if err != nil {
		return err
	}

	return store.SaveGroup(group)
}

func DeleteGroup(ctx context.Context, name string) error {
	store := ctx.Value("store").(Store)
	group, err := store.FindGroupByName(name)
	if err != nil {
		return err
	}
	return store.DeleteGroup(group)
}
