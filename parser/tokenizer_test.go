package parser

import "testing"

func getResult(text string) *TokenResult {
	tokenizer := &Tokenizer{}
	result, _ := tokenizer.Tokenize(text)
	return result
}

func TestTokenize(t *testing.T) {
	result := getResult("This needs to be done, promptly. @josler, #foo, #bar https://josler.io")
	if len(result.Mentions) != 1 && result.Mentions[0] != "@josler" {
		t.Errorf("mentions not parsed correctly")
	}
	if len(result.Hashtags) != 2 && result.Hashtags[0] != "#foo" && result.Hashtags[1] != "#bar" {
		t.Errorf("hashtags not parsed correctly")
	}
	if len(result.URLs) != 1 && result.URLs[0] != "https://josler.io" {
		t.Errorf("urls not parsed correctly")
	}
	if result.Stripped != "This needs to be done, promptly." {
		t.Errorf("text not stripped correctly")
	}
	if result.Raw != "This needs to be done, promptly. @josler, #foo, #bar https://josler.io" {
		t.Errorf("raw text altered!")
	}
}
