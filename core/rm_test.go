package core

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRm(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "2018-04-02")
	found := mostRecentItem(store)
	err := Rm(ctx, found.ID())
	if err != nil {
		t.Errorf("item not removed")
	}
	err = Show(ctx, found.ID())
	if err == nil {
		t.Errorf("item not removed")
	}
}

func TestRmMultiMatching(t *testing.T) {
	ctx, store := contextWithMemoryStore()
	Add(ctx, strings.NewReader("my new item"), "2018-04-02")
	found := mostRecentItem(store)

	// second item
	item := NewItem("will have similar id", time.Now())
	item.SetID(fmt.Sprintf("%s%s", found.ID()[:3], "yyy"))
	err := store.Save(item)
	if err != nil {
		t.Errorf("failed to save!")
	}

	err = Rm(ctx, found.ID()[:3])
	if err == nil {
		t.Errorf("rm didnt error as it should")
	}
}
