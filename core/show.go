package core

import (
	"context"
)

func Show(ctx context.Context, idString string, showConnected bool) error {
	items, err := FindAll(ctx, idString)
	if err != nil {
		return err
	}

	if len(items) == 1 && showConnected {
		connectedItems := getValidConnections(ctx, items[0])
		if len(connectedItems) > 0 {
			NewItemPrinter(ctx).PrintSingleWithConnected(items[0], connectedItems...)
			return nil
		}
	}
	NewItemPrinter(ctx).Print(items...)
	return nil
}

func getValidConnections(ctx context.Context, item *Item) []*Item {
	filteredConnections := []*Item{}
	for _, connection := range item.Connections() {
		found, err := FindOneOrPrint(ctx, connection)
		if err != nil {
			continue // invalid connection
		}
		filteredConnections = append(filteredConnections, found)
	}
	return filteredConnections
}
