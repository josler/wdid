package core

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestEdit(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item"), "2018-04-02")
		found := mostRecentItem(store)
		timespan, _ := TimeParser{Input: "2018-04-07"}.Parse()
		Edit(ctx, found.ID(), strings.NewReader("change the message"), "2018-04-07")
		found, err := store.Find(found.ID())
		if err != nil || found.Data() != "change the message" {
			t.Errorf("item not edited")
		}
		if err != nil || found.Time().Unix() != timespan.Start.Unix() {
			t.Errorf("item not edited %d", found.Time().Unix())
		}
	})
}

func TestEditAllowsFutureItem(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item"), "2018-04-02")
		found := mostRecentItem(store)
		err := Edit(ctx, found.ID(), strings.NewReader("change the message"), "2025-01-01")
		if err != nil {
			t.Errorf("item not edited")
		}
	})
}

func TestEditWithFunkyDateFails(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("standard item"), "2018-04-02")
		found := mostRecentItem(store)
		err := Edit(ctx, found.ID(), strings.NewReader("change the message"), "2019:10:10")
		assert.Error(t, err, "failed to parse time with input: 2019:10:10")
	})
}

func TestEditTrimsNewlines(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item"), "2018-04-02")
		found := mostRecentItem(store)
		err := Edit(ctx, found.ID(), strings.NewReader("change the message\n\n\n"), "2019-01-01")
		if err != nil {
			t.Errorf("item not edited")
		}
		found, err = store.Find(found.ID())
		assert.NilError(t, err)
		assert.Equal(t, found.Data(), "change the message", "doesn't trim newlines correctly")
	})
}
