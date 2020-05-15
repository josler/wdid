package core

import (
	"context"
)

func Show(ctx context.Context, idString string, showConnected bool) error {
	items, err := FindAll(ctx, idString)
	if err != nil {
		return err
	}
	printFormat := GetPrintFormatFromContext(ctx)
	shouldShowConnected := (showConnected || printFormat == HumanPrintFormat)

	if len(items) == 1 && shouldShowConnected {
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
	store := ctx.Value("store").(Store)
	filteredConnections := []*Item{}
	for _, connection := range item.Connections() {
		found, err := store.Find(connection)
		if err != nil {
			continue // invalid connection
		}
		filteredConnections = append(filteredConnections, found)
	}
	return filteredConnections
}
