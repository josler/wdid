package core

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestFindAll(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item to find"), "2018-04-02")
		found := mostRecentItem(store)

		items, err := FindAll(ctx, found.ID())
		assert.NilError(t, err)
		// in this case we don't have duplicates
		assert.Equal(t, len(items), 1)
		assert.Equal(t, items[0].Data(), "my new item to find")
	})
}
