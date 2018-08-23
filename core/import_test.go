package core_test

import (
	"bytes"
	"context"
	"testing"

	"gitlab.com/josler/wdid/core"
)

func contextWithBoltStore() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "store", boltStore)
}

func TestImport(t *testing.T) {
	withFreshDB(func() {
		ctx := contextWithBoltStore()
		f := bytes.NewBufferString("s36i4z	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, f)
		found, err := boltStore.Find("s36i4z")
		if err != nil || found.Data() != "some change" {
			t.Errorf("item not saved")
		}
	})
}

func TestImportExistingSameInternalID(t *testing.T) {
	withFreshDB(func() {
		ctx := contextWithBoltStore()
		f := bytes.NewBufferString("s36i4c	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, f)
		g := bytes.NewBufferString("s36i4c	recEJFQBuZsArxrJI	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, g)

		found, err := boltStore.Find("s36i4c")
		if err != nil {
			t.Fatalf("item not saved")
		}
		if found.Data() != "more detail" {
			t.Errorf("item has wrong data %s", found.Data())
		}
	})
}

func TestImportExistingDifferentInternalID(t *testing.T) {
	withFreshDB(func() {
		ctx := contextWithBoltStore()
		f := bytes.NewBufferString("s36i4b	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, f)
		g := bytes.NewBufferString("s36i4b	zzzzzz	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, g)

		found, err := boltStore.Find("s36i4b")
		if err != nil || found.Data() != "more detail" {
			t.Errorf("item not saved")
		}
	})
}

func TestImportExistingNoInternalID(t *testing.T) {
	withFreshDB(func() {
		ctx := contextWithBoltStore()
		f := bytes.NewBufferString("s36i4b	recEJFQBuZsArxrJI	done	<-4agi3u	some change	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, f)
		g := bytes.NewBufferString("s36i4b	 	done	<-4agi3u	more detail	2018-04-11T08:15:00-04:00")
		core.ReadToStore(ctx, g)

		found, err := boltStore.Find("s36i4b")
		if err != nil || found.Data() != "more detail" {
			t.Errorf("item not saved")
		}
	})
}
