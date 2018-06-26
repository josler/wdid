package core

import (
	"strings"
	"testing"
)

func TestEdit(t *testing.T) {
	ctx, store := contextWithMemoryStore()
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
}

func TestEditAllowsFutureItem(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "2018-04-02")
	found := mostRecentItem(store)
	err := Edit(ctx, found.ID(), strings.NewReader("change the message"), "2025-01-01")
	if err != nil {
		t.Errorf("item not edited")
	}
}
