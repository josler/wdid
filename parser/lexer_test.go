package parser

import (
	"testing"
)

func TestLexerBasic(t *testing.T) {
	_, itemchan := lex("tag=@josler")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 4 {
		t.Errorf("failed to lex correct number of items")
	}
	assertLexedItemTypeValue(t, lexItems[0], lexItemIdentifier, "tag")
	assertLexedItemTypeValue(t, lexItems[1], lexItemEq, "=")
	assertLexedItemTypeValue(t, lexItems[2], lexItemString, "@josler")
}

func TestLexerMultiple(t *testing.T) {
	_, itemchan := lex("tag=@josler,tag=#hashtag")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 8 {
		t.Errorf("failed to lex correct number of items")
	}
	assertLexedItemTypeValue(t, lexItems[3], lexItemComma, ",")
	assertLexedItemTypeValue(t, lexItems[4], lexItemIdentifier, "tag")
	assertLexedItemTypeValue(t, lexItems[5], lexItemEq, "=")
	assertLexedItemTypeValue(t, lexItems[6], lexItemString, "#hashtag")
}

func TestLexerNotEqual(t *testing.T) {
	_, itemchan := lex("tag!=@josler,tag=#hashtag")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 8 {
		t.Errorf("failed to lex correct number of items")
	}
	assertLexedItemTypeValue(t, lexItems[0], lexItemIdentifier, "tag")
	assertLexedItemTypeValue(t, lexItems[1], lexItemNe, "!=")
	assertLexedItemTypeValue(t, lexItems[2], lexItemString, "@josler")
	assertLexedItemTypeValue(t, lexItems[3], lexItemComma, ",")
	assertLexedItemTypeValue(t, lexItems[4], lexItemIdentifier, "tag")
	assertLexedItemTypeValue(t, lexItems[5], lexItemEq, "=")
	assertLexedItemTypeValue(t, lexItems[6], lexItemString, "#hashtag")
}

func TestLexerSpaces(t *testing.T) {
	_, itemchan := lex("tag=my tag")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 4 {
		t.Errorf("failed to lex correct number of items")
	}
	assertLexedItemTypeValue(t, lexItems[0], lexItemIdentifier, "tag")
	assertLexedItemTypeValue(t, lexItems[1], lexItemEq, "=")
	assertLexedItemTypeValue(t, lexItems[2], lexItemString, "my tag")
}

func TestLexerMissingValue(t *testing.T) {
	_, itemchan := lex("tag=")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 3 {
		t.Errorf("failed to lex correct number of items")
	}
	if lexItems[2].typ != lexItemEOF {
		t.Errorf("failed to correctly lex, expected: EOF")
	}
}

func TestLexerMissingValueComma(t *testing.T) {
	_, itemchan := lex("tag=,")
	lexItems := drainLexedItems(itemchan)
	if len(lexItems) != 4 {
		t.Errorf("failed to lex correct number of items")
	}
	if lexItems[3].typ != lexItemEOF {
		t.Errorf("failed to correctly lex, expected: EOF")
	}
}

func drainLexedItems(itemchan chan lexedItem) []lexedItem {
	lexItems := []lexedItem{}
	for i := range itemchan {
		lexItems = append(lexItems, i)
	}
	return lexItems
}

func assertLexedItemTypeValue(t *testing.T, item lexedItem, typ lexedItemType, val string) {
	if item.val != val || item.typ != typ {
		t.Errorf("failed to correctly lex, expected: %q, got: %v", val, item)
	}
}
