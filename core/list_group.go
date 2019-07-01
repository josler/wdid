package core

import (
	"context"
	"fmt"
)

func ListGroup(ctx context.Context) error {
	store := ctx.Value("store").(Store)
	groups, err := store.ListGroups()
	for _, group := range groups {
		fmt.Println(group)
	}
	return err
}
