package core

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/search/query"
	"gitlab.com/josler/wdid/filter"
)

type BleveItemDocument struct {
	ItemID   string
	Data     string
	Status   string
	Datetime int64
	Tag      *BleveTagDocument
}

type BleveTagDocument struct {
	TagID string
	Name  string
}

func CreateBleveIndex(indexname string, memory bool) (bleve.Index, error) {
	// create a mapping that's an item with a nested tag
	indexMapping := bleve.NewIndexMapping()
	itemMapping := bleve.NewDocumentMapping()
	indexMapping.AddDocumentMapping("item", itemMapping)

	dataFieldMapping := bleve.NewTextFieldMapping()
	dataFieldMapping.Analyzer = "en"
	itemMapping.AddFieldMappingsAt("Data", dataFieldMapping)

	statusFieldMapping := bleve.NewTextFieldMapping()
	itemMapping.AddFieldMappingsAt("Status", statusFieldMapping)

	createdAtFieldMapping := bleve.NewNumericFieldMapping()
	itemMapping.AddFieldMappingsAt("-Datetime", createdAtFieldMapping)

	tagDocumentMapping := bleve.NewDocumentMapping()
	tagNameFieldMapping := bleve.NewTextFieldMapping()
	tagNameFieldMapping.Analyzer = keyword.Name
	tagNameFieldMapping.IncludeTermVectors = false
	tagDocumentMapping.AddFieldMappingsAt("Name", tagNameFieldMapping)

	itemMapping.AddSubDocumentMapping("Tag", tagDocumentMapping)

	if memory {
		return bleve.NewMemOnly(indexMapping)
	}

	index, err := bleve.New(indexname, indexMapping)
	if err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(indexname)
	}
	return index, err
}

func SaveBleve(index bleve.Index, store Store, item *Item) {
	bleveItem := &BleveItemDocument{
		ItemID:   item.ID(),
		Data:     item.Data(),
		Status:   item.Status(),
		Datetime: item.Time().Unix(),
	}
	tags := item.Tags()
	if len(tags) == 0 {
		// no tags, save as-is
		err := index.Index(fmt.Sprintf("%s", bleveItem.ItemID), bleveItem)
		if err != nil {
			panic(err)
		}
	}

	// save variation per-tag
	for _, tag := range item.Tags() {
		found, err := store.FindTag(tag.Name())
		if err != nil {
			// might not have a tag yet
			continue
		}
		bleveItem.Tag = &BleveTagDocument{
			TagID: found.internalID,
			Name:  found.Name(),
		}
		id := fmt.Sprintf("%s:%s", bleveItem.ItemID, bleveItem.Tag.TagID)
		err = index.Index(id, bleveItem)
		if err != nil {
			panic(err)
		}
	}
}

func Delete(index bleve.Index, store Store, item *Item) error {
	tags := item.Tags()
	if len(tags) == 0 {
		// no tags, save as-is
		err := index.Delete(fmt.Sprintf("%s", item.ID()))
		if err != nil {
			return err
		}
	}

	// delete variation per-tag
	for _, tag := range item.Tags() {
		found, _ := store.FindTag(tag.Name())
		id := fmt.Sprintf("%s:%s", item.ID(), found.internalID)
		err := index.Delete(id)
		if err != nil {
			return err
		}
	}

	return nil
}

func Query(index bleve.Index, filters ...filter.Filter) ([]*Item, error) {
	items := []*Item{}
	queryFilters := []query.Query{}
	for _, filter := range filters {
		q, err := filter.BleveQuery()
		if err != nil {
			return items, err
		}
		queryFilters = append(queryFilters, q)
	}

	// do search
	query := query.NewBooleanQuery(queryFilters, nil, nil) // must, should, mustNot
	search := bleve.NewSearchRequest(query)
	search.Size = 100000000 // Big number, aka all results
	search.Fields = []string{"ItemID", "Data", "Datetime", "Status"}
	search.SortBy([]string{"Datetime"})
	searchResults, err := index.Search(search)
	if err != nil {
		return items, err
	}

	foundMap := map[string]struct{}{}

	for _, hit := range searchResults.Hits {
		id := fmt.Sprintf("%v", hit.Fields["ItemID"])
		if _, alreadyGot := foundMap[id]; alreadyGot {
			continue // already have this
		}

		data := fmt.Sprintf("%v", hit.Fields["Data"])
		status := fmt.Sprintf("%v", hit.Fields["Status"])

		var at time.Time
		switch i := hit.Fields["Datetime"].(type) {
		case float64:
			at = time.Unix(int64(i), 0)
		}
		item := NewItem(data, at)
		item.SetID(id)
		item.status = status

		foundMap[id] = struct{}{}
		items = append(items, item)
	}
	return items, err
}
