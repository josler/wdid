package core_test

import (
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"gitlab.com/josler/wdid/core"
)

var boltStore *core.BoltStore

func TestMain(m *testing.M) {
	db, err := storm.Open("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}
	boltStore = core.NewBoltStore(db)
	result := m.Run()

	db.Close()
	os.Exit(result)
}

func withFreshDB(f func()) {
	boltStore.DropBucket("StormItem")
	f()
}

func TestSaveAlreadyExists(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		err := boltStore.Save(item)
		if err != nil {
			t.Fatalf("error %s", err)
		}
		item.ResetInternalID()
		err = boltStore.Save(item)
		if err != nil && err != storm.ErrAlreadyExists {
			t.Fatalf("error %s", err)
		}
	})
}

func TestSaveUpdate(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		err := boltStore.Save(item)
		if err != nil {
			t.Fatalf("error %s", err)
		}
		item.Do()
		err = boltStore.Save(item)
		if err != nil || item.Status() != "done" {
			t.Fatalf("error updating item")
		}
	})
}

func TestList(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now().Add(-1*time.Minute))
		err := boltStore.Save(item)
		if err != nil {
			t.Fatalf("error %s", err)
		}
		items, _ := boltStore.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()))
		if len(items) != 1 {
			t.Fatalf("error: no items found")
		}
		if items[0].ID() != item.ID() {
			t.Errorf("error id not matching")
		}
	})
}
func TestListDate(t *testing.T) {
	withFreshDB(func() {
		now := time.Now()
		boltStore.Save(core.NewItem("1", now.Add(-48*time.Hour)))
		boltStore.Save(core.NewItem("2", now.Add(-24*time.Hour)))
		boltStore.Save(core.NewItem("3", now.Add(-1*time.Minute)))
		boltStore.Save(core.NewItem("4", now.Add(24*time.Hour)))
		boltStore.Save(core.NewItem("5", now.Add(1*time.Second))) // should not pick this up as it's greater than end time

		items, _ := boltStore.List(core.NewTimespan(now.Add(-36*time.Hour), now))
		if len(items) != 2 {
			t.Fatalf("error: not all items found")
		}
		if items[0].Data() != "2" {
			t.Errorf("error data not matching")
		}
		if items[1].Data() != "3" {
			t.Errorf("error data not matching")
		}
	})
}

func TestListStatus(t *testing.T) {
	withFreshDB(func() {
		boltStore.Save(core.NewItem("1", time.Now()))
		doneItem := core.NewItem("2", time.Now())
		doneItem.Do()
		boltStore.Save(doneItem)
		skippedItem := core.NewItem("3", time.Now())
		skippedItem.Skip()
		boltStore.Save(skippedItem)

		items, _ := boltStore.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()), "waiting", "skipped")
		if len(items) != 2 {
			t.Fatalf("error: not all items found")
		}
		if items[0].Data() != "1" {
			t.Errorf("error data not matching")
		}
		if items[1].Data() != "3" {
			t.Errorf("error data not matching")
		}

		items, _ = boltStore.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()), "done")
		if len(items) != 1 {
			t.Fatalf("error: not all items found")
		}
		if items[0].Data() != "2" {
			t.Errorf("error data not matching")
		}
	})
}

func TestShow(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		boltStore.Save(item)
		found, err := boltStore.Find(item.ID())
		if err != nil || found.ID() != item.ID() {
			t.Errorf("error item not found correctly")
		}
	})
}

func TestShowPartialID(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		boltStore.Save(item)
		found, err := boltStore.Find(item.ID()[:2])
		if err != nil || found.ID() != item.ID() {
			t.Errorf("error item not found correctly")
		}
	})
}

func TestDelete(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		boltStore.Save(item)
		boltStore.Delete(item)
		_, err := boltStore.Find(item.ID())
		if err != storm.ErrNotFound {
			t.Errorf("error item not found correctly")
		}
	})
}
