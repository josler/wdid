package migrations

import (
	"context"

	"github.com/josler/wdid/core"
)

func AddKinds(ctx context.Context, from string) {
	items, err := core.ListWithoutPrinting(ctx, from)

	store := ctx.Value("store").(core.Store)

	if err != nil {
		panic(err)
	}

	for _, item := range items {
		if item.Kind() != core.Task {
			continue
		}

		tags := item.Tags()
		for _, tag := range tags {
			if tag.Name() == "#note" {
				item.Do()
				item.SetKind(core.Note)
				err := store.Save(item)
				if err != nil {
					panic(err)
				}
			}
		}

	}

}
