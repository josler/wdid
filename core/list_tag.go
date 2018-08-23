package core

import (
	"context"
	"fmt"
)

func ListTag(ctx context.Context) error {
	store := ctx.Value("store").(Store)
	tags, err := store.ListTags()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		fmt.Println(tag)
	}
	return nil
}
