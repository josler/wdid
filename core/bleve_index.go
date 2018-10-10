package core

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"gitlab.com/josler/wdid/filter"
)

type BleveItemDocument struct {
	ItemID   string
	Data     string
	Status   string
	Datetime int64
	TagIDs   []string
}

func CreateBleveIndex(indexname string, memory bool) (bleve.Index, error) {
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

	// let TagIDs be default mapping

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
	bleveItemSingle := &BleveItemDocument{
		ItemID:   item.ID(),
		Data:     item.Data(),
		Status:   item.Status(),
		Datetime: item.Time().Unix(),
	}

	// load TagIDs
	for _, tag := range item.Tags() {
		found, err := store.FindTag(tag.Name())
		if err != nil {
			// might not have a tag record in db
			continue
		}
		bleveItemSingle.TagIDs = append(bleveItemSingle.TagIDs, found.internalID)
	}

	err := index.Index(fmt.Sprintf("%s", bleveItemSingle.ItemID), bleveItemSingle)
	if err != nil {
		panic(err)
	}
}

func Delete(index bleve.Index, store Store, item *Item) error {
	return index.Delete(fmt.Sprintf("%s", item.ID()))
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
