package core_test

import (
	"fmt"
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
	boltStore.DropBucket("StormTag")
	boltStore.DropBucket("StormItemTag")
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

func TestFind(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("some data", time.Now())
		boltStore.Save(item)
		found, err := boltStore.Find(item.ID())
		if err != nil || found.ID() != item.ID() {
			t.Errorf("error item not found correctly")
		}
	})
}

func TestFindAll(t *testing.T) {
	withFreshDB(func() {
		item := core.NewItem("to be saved twice", time.Now())
		boltStore.Save(item)
		item.ResetInternalID()
		item.SetID(fmt.Sprintf("%s%s", item.ID()[:3], "yyy"))
		err := boltStore.Save(item) // save a copy
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		found, err := boltStore.FindAll(item.ID()[:2])
		if err != nil || len(found) != 2 {
			t.Errorf("error items not found correctly")
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

func TestSaveTag(t *testing.T) {
	withFreshDB(func() {
		tag := core.NewTag("mytag")
		err := boltStore.SaveTag(tag)
		if err != nil {
			t.Errorf("failed to save tag")
		}
	})
}

func TestFindTag(t *testing.T) {
	withFreshDB(func() {
		tag := core.NewTag("mytag")
		err := boltStore.SaveTag(tag)
		if err != nil {
			t.Errorf("failed to save tag")
		}
		found, err := boltStore.FindTag("mytag")
		if err != nil || found == nil || found.Name() != "mytag" {
			t.Errorf("failed to find tag")
		}
	})
}

func TestListTags(t *testing.T) {
	withFreshDB(func() {
		tagone := core.NewTag("one")
		boltStore.SaveTag(tagone)
		tagtwo := core.NewTag("two")
		boltStore.SaveTag(tagtwo)

		found, err := boltStore.ListTags()
		if err != nil || len(found) != 2 {
			t.Errorf("failed to list tags")
		}

		if found[0].Name() != tagone.Name() || found[1].Name() != tagtwo.Name() {
			t.Errorf("failed to list tags in order")
		}

	})
}

func TestSaveItemTag(t *testing.T) {
	withFreshDB(func() {
		tag := core.NewTag("mytag")
		item := core.NewItem("my item", time.Now())

		err := boltStore.SaveItemTag(item, tag)
		if err != nil {
			t.Errorf("failed to save item tag")
		}
	})
}

func TestDeleteItemTag(t *testing.T) {
	withFreshDB(func() {
		tag := core.NewTag("mytag")
		item := core.NewItem("my item", time.Now())

		err := boltStore.SaveItemTag(item, tag)
		if err != nil {
			t.Errorf("failed to save item tag")
		}
		err = boltStore.DeleteItemTag(item, tag)
		if err != nil {
			t.Errorf("failed to delete item tag")
		}
	})
}

func TestFindItemsWithTag(t *testing.T) {
	withFreshDB(func() {
		tag := core.NewTag("mytag")
		boltStore.SaveTag(tag)
		item := core.NewItem("my item", time.Now())
		boltStore.Save(item)
		itemtwo := core.NewItem("my second item", time.Now())
		boltStore.Save(itemtwo)
		itemthree := core.NewItem("my third item", time.Now())
		boltStore.Save(itemthree)

		boltStore.SaveItemTag(item, tag)
		boltStore.SaveItemTag(itemtwo, tag)

		items, err := boltStore.FindItemsWithTag(tag)
		if err != nil || len(items) != 2 {
			t.Errorf("failed to find items with tag")
		}
		if items[0].Data() != "my item" {
			t.Errorf("found wrong item through tag")
		}
		if items[1].Data() != "my second item" {
			t.Errorf("found wrong item through tag")
		}
	})
}
