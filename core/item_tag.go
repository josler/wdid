package core

import "time"

type ItemTag struct {
	itemID    string
	tagID     string
	createdAt time.Time
}

func NewItemTag(item *Item, tag *Tag) *ItemTag {
	return &ItemTag{itemID: item.ID(), tagID: tag.internalID}
}

func (it *ItemTag) ItemID() string {
	return it.itemID
}

func (it *ItemTag) TagID() string {
	return it.tagID
}

func (it *ItemTag) CreatedAt() time.Time {
	return it.createdAt
}
