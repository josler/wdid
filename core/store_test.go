package core_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/josler/wdid/core"
	"github.com/josler/wdid/filter"
)

func contextWithStore(store core.Store) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "store", store)
}

func withFreshBoltStore(boltStore *core.BoltStore, f func()) {
	boltStore.DropBucket("StormItem")
	boltStore.DropBucket("StormTag")
	boltStore.DropBucket("StormItemTag")
	boltStore.DropBucket("StormGroup")
	f()
}

type storeTest func(t *testing.T, store core.Store)

// subtests for the store
func tests() map[string]storeTest {
	return map[string]storeTest{
		"saveAlreadyExists":             saveAlreadyExists,
		"saveUpdate":                    saveUpdate,
		"list":                          list,
		"listEmptyShouldNotError":       listEmptyShouldNotError,
		"listDate":                      listDate,
		"listStatus":                    listStatus,
		"listFilters":                   listFilters,
		"listFiltersNe":                 listFiltersNe,
		"listFiltersStatusOr":           listFiltersStatusOr,
		"listFiltersGroup":              listFiltersGroup,
		"listFiltersGroupNe":            listFiltersGroupNe,
		"find":                          find,
		"findMultipleReturnsMostRecent": findMultipleReturnsMostRecent,
		"findAll":                       findAll,
		"showPartialID":                 showPartialID,
		"doDelete":                      doDelete,
		"saveTag":                       saveTag,
		"findTag":                       findTag,
		"listTags":                      listTags,
		"saveItemTag":                   saveItemTag,
		"deleteItemTag":                 deleteItemTag,
		"findItemsWithTag":              findItemsWithTag,
		"deleteItemTagsWithItem":        deleteItemTagsWithItem,
		"saveGroup":                     saveGroup,
		"deleteGroup":                   deleteGroup,
		"listGroups":                    listGroups,
	}
}

func TestBoltStore(t *testing.T) {
	db, err := storm.Open("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}
	boltStore := core.NewBoltStore(db)

	for name, subTest := range tests() {
		t.Run(name, func(t *testing.T) {
			withFreshBoltStore(boltStore, func() {
				subTest(t, boltStore)
			})
		})
	}

	db.Close()
}

func saveAlreadyExists(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now())
	err := store.Save(item)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	item.ResetInternalID()
	err = store.Save(item)
	if err != nil && err != storm.ErrAlreadyExists {
		t.Fatalf("error %s", err)
	}
}

func saveUpdate(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now())
	err := store.Save(item)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	item.Do()
	err = store.Save(item)
	if err != nil || item.Status() != core.DoneStatus {
		t.Fatalf("error updating item")
	}
}

func list(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now().Add(-1*time.Minute))
	err := store.Save(item)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	items, _ := store.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()))
	if len(items) != 1 {
		t.Fatalf("error: no items found")
	}
	if items[0].ID() != item.ID() {
		t.Errorf("error id not matching")
	}
}

func listEmptyShouldNotError(t *testing.T, store core.Store) {
	// interestingly, this test doesn't fail when the database is completely empty
	// it only fails when there has been at least one write.
	item := core.NewItem("some data", time.Now().Add(-1*time.Minute))
	store.Save(item)
	items, err := store.List(core.NewTimespan(time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour)))
	if len(items) != 0 {
		t.Fatalf("error: items found when they shouldn't have been")
	}
	if err != nil {
		t.Errorf("error returned for empty list, %v", err)
	}
}

func listDate(t *testing.T, store core.Store) {
	now := time.Now()
	store.Save(core.NewItem("1", now.Add(-48*time.Hour)))
	store.Save(core.NewItem("2", now.Add(-24*time.Hour)))
	store.Save(core.NewItem("3", now.Add(-1*time.Minute)))
	store.Save(core.NewItem("4", now.Add(24*time.Hour)))
	store.Save(core.NewItem("5", now.Add(1*time.Second))) // should not pick this up as it's greater than end time

	items, _ := store.List(core.NewTimespan(now.Add(-36*time.Hour), now))
	if len(items) != 2 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "2" {
		t.Errorf("error data not matching")
	}
	if items[1].Data() != "3" {
		t.Errorf("error data not matching")
	}
}

func listStatus(t *testing.T, store core.Store) {
	store.Save(core.NewItem("1", time.Now()))
	doneItem := core.NewItem("2", time.Now())
	doneItem.Do()
	store.Save(doneItem)
	skippedItem := core.NewItem("3", time.Now())
	skippedItem.Skip()
	store.Save(skippedItem)

	items, _ := store.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()), core.WaitingStatus, core.SkippedStatus)
	if len(items) != 2 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "1" {
		t.Errorf("error data not matching")
	}
	if items[1].Data() != "3" {
		t.Errorf("error data not matching")
	}

	items, _ = store.List(core.NewTimespan(time.Now().Add(-1*time.Hour), time.Now()), core.DoneStatus)
	if len(items) != 1 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "2" {
		t.Errorf("error data not matching")
	}
}

func setupTagAndItems(store core.Store) {
	tag := core.NewTag("#mytag")
	store.SaveTag(tag)

	item := core.NewItem("my item", time.Now())
	store.Save(item)
	doneItem := core.NewItem("#mytag done", time.Now())
	doneItem.Do()
	store.Save(doneItem)
	skippedItem := core.NewItem("#mytag skipped", time.Now())
	skippedItem.Skip()
	store.Save(skippedItem)

	store.SaveItemTag(item, tag)
	store.SaveItemTag(skippedItem, tag)
}

func listFilters(t *testing.T, store core.Store) {
	setupTagAndItems(store)

	filters := []filter.Filter{
		core.NewStatusFilter(filter.FilterEq, "skipped"),
		core.NewTagFilter(store, filter.FilterEq, "#mytag"),
	}
	items, _ := store.ListFilters(filters)

	if len(items) != 1 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "#mytag skipped" {
		t.Errorf("data not matching")
	}
}

func listFiltersNe(t *testing.T, store core.Store) {
	setupTagAndItems(store)

	filters := []filter.Filter{
		core.NewTagFilter(store, filter.FilterNe, "#mytag"),
	}
	items, _ := store.ListFilters(filters)

	if len(items) != 1 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "my item" {
		t.Errorf("data not matching")
	}
}

func listFiltersStatusOr(t *testing.T, store core.Store) {
	setupTagAndItems(store)

	filters := []filter.Filter{
		core.NewStatusFilter(filter.FilterEq, "skipped", "done"),
	}
	items, _ := store.ListFilters(filters)

	if len(items) != 2 {
		t.Fatalf("error: not all items found")
	}
	if items[0].Data() != "#mytag done" {
		t.Errorf("data not matching")
	}

	if items[1].Data() != "#mytag skipped" {
		t.Errorf("data not matching")
	}
}

func listFiltersGroup(t *testing.T, store core.Store) {
	setupTagAndItems(store)

	filters := []filter.Filter{
		core.NewGroupFilter(filter.FilterEq, "name", []filter.Filter{
			core.NewTagFilter(store, filter.FilterEq, "#mytag"),
			core.NewStatusFilter(filter.FilterNe, "done"),
		}),
	}

	items, _ := store.ListFilters(filters)
	if len(items) != 1 {
		t.Fatalf("wrong items found %v", items)
	}

	if items[0].Data() != "#mytag skipped" {
		t.Errorf("data not matching")
	}
}

func listFiltersGroupNe(t *testing.T, store core.Store) {
	setupTagAndItems(store)

	filters := []filter.Filter{
		core.NewGroupFilter(filter.FilterNe, "name", []filter.Filter{
			core.NewTagFilter(store, filter.FilterEq, "#mytag"),
		}),
	}

	items, _ := store.ListFilters(filters)
	if len(items) != 1 {
		t.Fatalf("wrong items found %v", items)
	}

	if items[0].Data() != "my item" {
		t.Errorf("data not matching")
	}
}

func find(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now())
	store.Save(item)
	found, err := store.Find(item.ID())
	if err != nil || found.ID() != item.ID() {
		t.Errorf("error item not found correctly")
	}
}

func findMultipleReturnsMostRecent(t *testing.T, store core.Store) {
	item := core.NewItem("to be saved twice", time.Now().Add(-5*time.Second))
	firstID := item.ID()
	store.Save(item)

	item = core.NewItem("to be saved twice", time.Now())
	item.SetID(fmt.Sprintf("%s%s", firstID[:3], "yyy"))
	err := store.Save(item) // save a copy
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	found, err := store.Find(item.ID()[:2])
	if err != nil {
		t.Errorf("error item not found correctly")
	}
	if found.ID() != item.ID() {
		t.Errorf("didnt return most recent item %s %s", found.ID(), item.ID())
	}
}

func findAll(t *testing.T, store core.Store) {
	item := core.NewItem("to be saved twice", time.Now())
	store.Save(item)
	item.ResetInternalID()
	item.SetID(fmt.Sprintf("%s%s", item.ID()[:3], "yyy"))
	err := store.Save(item) // save a copy
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	found, err := store.FindAll(item.ID()[:2])
	if err != nil || len(found) != 2 {
		t.Errorf("error items not found correctly")
	}
}

func showPartialID(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now())
	store.Save(item)
	found, err := store.Find(item.ID()[:2])
	if err != nil || found.ID() != item.ID() {
		t.Errorf("error item not found correctly")
	}
}

func doDelete(t *testing.T, store core.Store) {
	item := core.NewItem("some data", time.Now())
	store.Save(item)
	store.Delete(item)
	_, err := store.Find(item.ID())
	if err != storm.ErrNotFound {
		t.Errorf("error item not found correctly")
	}
}

func saveTag(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	err := store.SaveTag(tag)
	if err != nil {
		t.Errorf("failed to save tag")
	}
	tag = core.NewTag("mytag")
	err = store.SaveTag(tag)
	if err != nil {
		t.Errorf("failed to save tag")
	}
}

func findTag(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	err := store.SaveTag(tag)
	if err != nil {
		t.Errorf("failed to save tag")
	}
	found, err := store.FindTag("mytag")
	if err != nil || found == nil || found.Name() != "mytag" {
		t.Errorf("failed to find tag")
	}
}

func listTags(t *testing.T, store core.Store) {
	tagone := core.NewTag("one")
	store.SaveTag(tagone)
	tagtwo := core.NewTag("two")
	store.SaveTag(tagtwo)

	found, err := store.ListTags()
	if err != nil || len(found) != 2 {
		t.Errorf("failed to list tags")
	}

	if found[0].Name() != tagone.Name() || found[1].Name() != tagtwo.Name() {
		t.Errorf("failed to list tags in order")
	}
}

func saveItemTag(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	item := core.NewItem("my item", time.Now())

	err := store.SaveItemTag(item, tag)
	if err != nil {
		t.Errorf("failed to save item tag")
	}

	err = store.SaveItemTag(item, tag)
	if err != nil {
		t.Errorf("failed to save duplicate")
	}
}

func deleteItemTag(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	item := core.NewItem("my item", time.Now())

	err := store.SaveItemTag(item, tag)
	if err != nil {
		t.Errorf("failed to save item tag")
	}
	err = store.DeleteItemTag(item, tag)
	if err != nil {
		t.Errorf("failed to delete item tag")
	}
}

func findItemsWithTag(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	store.SaveTag(tag)
	item := core.NewItem("my item", time.Now())
	store.Save(item)
	itemtwo := core.NewItem("my second item", time.Now())
	store.Save(itemtwo)
	itemthree := core.NewItem("my third item", time.Now())
	store.Save(itemthree)

	store.SaveItemTag(item, tag)
	store.SaveItemTag(itemtwo, tag)

	items, err := store.FindItemsWithTag(tag, -1)
	if err != nil || len(items) != 2 {
		t.Errorf("failed to find items with tag")
	}
	if items[0].Data() != "my item" {
		t.Errorf("found wrong item through tag")
	}
	if items[1].Data() != "my second item" {
		t.Errorf("found wrong item through tag")
	}
}

func deleteItemTagsWithItem(t *testing.T, store core.Store) {
	tag := core.NewTag("mytag")
	item := core.NewItem("my item", time.Now())
	store.SaveTag(tag)

	err := store.SaveItemTag(item, tag)
	if err != nil {
		t.Errorf("failed to save item tag")
	}
	err = store.DeleteItemTagsWithItem(item)
	if err != nil {
		t.Errorf("failed to delete item tag")
	}
	items, err := store.FindItemsWithTag(tag, -1)
	if err != nil && len(items) != 0 {
		t.Errorf("failed to delete all item tags!")
	}
}

func saveGroup(t *testing.T, store core.Store) {
	group := core.NewGroup("group name", "tag=#foo,status!=done")
	err := store.SaveGroup(group)
	if err != nil {
		t.Fatalf("failed to save group %v", err)
	}

	group, err = store.FindGroupByName("group name")
	if err != nil {
		t.Errorf("failed to find group")
	}
	if group.FilterString != "tag=#foo,status!=done" {
		t.Error("failed to load group filterstring")
	}
}

func deleteGroup(t *testing.T, store core.Store) {
	group := core.NewGroup("group name", "tag=#foo,status!=done")
	err := store.SaveGroup(group)
	if err != nil {
		t.Fatalf("failed to save group %v", err)
	}

	err = store.DeleteGroup(group)
	if err != nil {
		t.Fatalf("failed to delete group %v", err)
	}
}

func listGroups(t *testing.T, store core.Store) {
	group := core.NewGroup("group name", "tag=#foo,status!=done")
	err := store.SaveGroup(group)
	if err != nil {
		t.Fatalf("failed to save group %v", err)
	}

	groups, err := store.ListGroups()
	if err != nil {
		t.Fatalf("failed to list groups %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("didn't load groups")
	}
	if groups[0].Name != "group name" {
		t.Errorf("did not load correct group")
	}
}
