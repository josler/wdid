package parser

import (
	"testing"
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
