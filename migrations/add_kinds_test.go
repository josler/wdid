package migrations

import (
	"context"
	"os"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/josler/wdid/core"
)

func contextWithStore(store core.Store) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "store", store)
}

func withFreshBoltStore(boltStore *core.BoltStore, f func()) {
	boltStore.DropBucket("StormItem")
	boltStore.DropBucket("StormTag")
	boltStore.DropBucket("StormGroup")
	f()
}

func TestAddKinds(t *testing.T) {
	store, err := core.NewBoltStore("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}

	withFreshBoltStore(store, func() {
		task := core.NewTask("some data", time.Now())
		assert.NilError(t, store.Save(task))

		note := core.NewNote("some data", time.Now())
		assert.NilError(t, store.Save(note))

		taskToNote := core.NewTask("some data #note", time.Now())
		assert.NilError(t, store.Save(taskToNote))

		ctx := contextWithStore(store)
		AddKinds(ctx, "10")
		assert.NilError(t, err)

		foundTask, _ := store.Find(task.ID())
		assert.Equal(t, foundTask.Kind(), core.Task)

		foundNote, _ := store.Find(note.ID())
		assert.Equal(t, foundNote.Kind(), core.Note)

		foundTaskToNote, _ := store.Find(taskToNote.ID())
		assert.Equal(t, foundTaskToNote.Kind(), core.Note)
		assert.Equal(t, foundTaskToNote.Status(), core.NoStatus)
	})

}
