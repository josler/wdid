package parser

import (
	"testing"
)

func getResult(text string) *TokenResult {
	tokenizer := &Tokenizer{}
	result, _ := tokenizer.Tokenize(text)
	return result
}

func TestTokenize(t *testing.T) {
	result := getResult("This needs to be done, promptly. @josler, #foo, #bar https://josler.io")
	if result.Tags[0] != "@josler" {
		t.Errorf("mentions not parsed correctly")
	}
	if result.Tags[1] != "#foo" && result.Tags[2] != "#bar" {
		t.Errorf("hashtags not parsed correctly")
	}
	if result.Raw != "This needs to be done, promptly. @josler, #foo, #bar https://josler.io" {
		t.Errorf("raw text altered!")
	}
}
