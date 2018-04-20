package core

import (
	"strings"
	"testing"
)

func TestEdit(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "2018-04-02")
	found := mostRecentItem(store)
	Edit(ctx, found.ID(), strings.NewReader("change the message"), "2018-04-07")
	found, err := store.Find(found.ID())
	if err != nil || found.Data() != "change the message" {
		t.Errorf("item not edited")
	}
	if err != nil || found.Time().Unix() != 1523073600 {
		t.Errorf("item not edited %d", found.Time().Unix())
	}
}

func TestEditPreventsFutureItem(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "2018-04-02")
	found := mostRecentItem(store)
	err := Edit(ctx, found.ID(), strings.NewReader("change the message"), "2025-01-01")
	if err == nil {
		t.Errorf("item was incorrectly saved")
	}
}
