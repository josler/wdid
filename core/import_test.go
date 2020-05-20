package core_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/josler/wdid/core"
)

func importTests() []storeTest {
	return []storeTest{
		testImport,
		testImportExistingSameInternalID,
		testImportExistingDifferentInternalID,
		testImportExistingNoInternalID,
		testImportKind,
	}
}

func TestBoltStoreImport(t *testing.T) {
	boltStore, err := core.NewBoltStore("/tmp/test123.db")
	if err != nil {
		os.Exit(1)
	}

	for _, test := range importTests() {
		withFreshBoltStore(boltStore, func() {
			test(t, boltStore)
		})
	}
}

func testImport(t *testing.T, store core.Store) {
	ctx := contextWithStore(store)
	f := bytes.NewBufferString("s36i4z	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, f)
	found, err := store.Find("s36i4z")
	if err != nil || found.Data() != "some change" {
		t.Errorf("item not saved")
	}
	if found.Kind() != core.Task {
		t.Errorf("did not save item as Task without explicit kind")
	}
}

func testImportExistingSameInternalID(t *testing.T, store core.Store) {
	ctx := contextWithStore(store)
	f := bytes.NewBufferString("s36i4c	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, f)
	g := bytes.NewBufferString("s36i4c	recEJFQBuZsArxrJI	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, g)

	found, err := store.Find("s36i4c")
	if err != nil {
		t.Fatalf("item not saved")
	}
	if found.Data() != "more detail" {
		t.Errorf("item has wrong data %s", found.Data())
	}
}

func testImportExistingDifferentInternalID(t *testing.T, store core.Store) {
	ctx := contextWithStore(store)
	f := bytes.NewBufferString("s36i4b	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, f)
	g := bytes.NewBufferString("s36i4b	zzzzzz	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, g)

	found, err := store.Find("s36i4b")
	if err != nil || found.Data() != "more detail" {
		t.Errorf("item not saved")
	}
}

func testImportExistingNoInternalID(t *testing.T, store core.Store) {
	ctx := contextWithStore(store)
	f := bytes.NewBufferString("s36i4b	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, f)
	g := bytes.NewBufferString("s36i4b	 	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
	core.ReadToStore(ctx, g)

	found, err := store.Find("s36i4b")
	if err != nil || found.Data() != "more detail" {
		t.Errorf("item not saved")
	}
}

func testImportKind(t *testing.T, store core.Store) {
	ctx := contextWithStore(store)
	f := bytes.NewBufferString("s36i4z	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00	note")
	core.ReadToStore(ctx, f)
	found, err := store.Find("s36i4z")
	if err != nil || found.Kind() != core.Note {
		t.Errorf("item not saved as Note")
	}
}
