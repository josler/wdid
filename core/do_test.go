package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestDoMultiMatching(t *testing.T) {
	contextWithStore(func(ctx context.Context, store Store) {
		Add(ctx, strings.NewReader("my new item"), "2018-04-02")
		found := mostRecentItem(store)

		// second item
		item := NewTask("will have similar id", time.Now())
		item.SetID(fmt.Sprintf("%s%s", found.ID()[:3], "yyy"))
		err := store.Save(item)
		assert.NilError(t, err, "failed to save!")

		err = Do(ctx, found.ID()[:3])
		assert.Error(t, err, "unable to find unique item", "Do didn't error as it should")
	})
}
