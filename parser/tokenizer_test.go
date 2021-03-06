package parser

import (
	"testing"

	"gotest.tools/assert"
)

func getResult(text string) *TokenResult {
	tokenizer := &Tokenizer{}
	result, _ := tokenizer.Tokenize(text)
	return result
}

func includes(slice []string, search string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}

func TestTokenize(t *testing.T) {
	result := getResult("This needs to be done, promptly. @josler, #foo, #bar https://josler.io")
	if !includes(result.Tags, "@josler") {
		t.Errorf("mentions not parsed correctly")
	}
	if !includes(result.Tags, "#foo") || !includes(result.Tags, "#bar") {
		t.Errorf("hashtags not parsed correctly")
	}
	if result.Raw != "This needs to be done, promptly. @josler, #foo, #bar https://josler.io" {
		t.Errorf("raw text altered!")
	}
}

func TestTokenizeStartEnd(t *testing.T) {
	result := getResult("#foo ##nope #bar")
	assert.DeepEqual(t, []string{"#foo", "#bar"}, result.Tags)

	result = getResult("@foo @@nope @bar")
	assert.DeepEqual(t, []string{"@foo", "@bar"}, result.Tags)
}

func TestTokenizeDoubleSquareBrackets(t *testing.T) {
	result := getResult("[[connection_to]] whatever [not conn]")
	assert.DeepEqual(t, []string{"connection_to"}, result.Connections)
}

func TestTokenizeDoubleSquareBracketsBroken(t *testing.T) {
	result := getResult("[[[connection_to]] whatever [not conn] [[realconn]] [[foo]bar] [[bax]] [[connection:title]]")
	assert.DeepEqual(t, []string{"connection_to", "realconn", "bax", "connection:title"}, result.Connections)
}

func TestTokenizeTextAfterConnections(t *testing.T) {
	result := getResult("[[connection_to]](comment)")
	assert.DeepEqual(t, []string{"connection_to"}, result.Connections)
}

func TestTokenizeDuplicates(t *testing.T) {
	result := getResult("This needs to be done, promptly. #foo #foo")
	if len(result.Tags) != 1 {
		t.Errorf("failed to parse duplicates")
	}
	if !includes(result.Tags, "#foo") {
		t.Errorf("hashtags not parsed correctly")
	}
}

func TestTokenizeMultiplePrefixes(t *testing.T) {
	result := getResult("This needs to be done, promptly. ##foo @@foo ## @@@")
	if len(result.Tags) != 0 {
		t.Errorf("failed to ignore double prefixes")
	}
}
