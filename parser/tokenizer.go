package parser

import (
	"strings"

	"gopkg.in/jdkato/prose.v2"
)

type TokenResult struct {
	Mentions []string
	Hashtags []string
	URLs     []string
	Stripped string
	Raw      string
}

type Tokenizer struct {
}

func (t *Tokenizer) Tokenize(text string) (*TokenResult, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, err
	}

	result := TokenResult{Mentions: []string{}, Hashtags: []string{}, URLs: []string{}}
	result.Raw = text

	for _, tok := range doc.Tokens() {
		if strings.HasPrefix(tok.Text, "@") {
			result.Mentions = append(result.Mentions, tok.Text)
		} else if strings.HasPrefix(tok.Text, "#") {
			result.Hashtags = append(result.Hashtags, tok.Text)
		} else if tok.Tag == "NN" && strings.HasPrefix(tok.Text, "https") {
			result.URLs = append(result.URLs, tok.Text)
		}
	}

	result.Stripped = result.Raw

	result.Stripped = t.strip(result.Stripped, result.Mentions)
	result.Stripped = t.strip(result.Stripped, result.Hashtags)
	result.Stripped = t.strip(result.Stripped, result.URLs)
	result.Stripped = strings.Trim(result.Stripped, " ,") // strip leading/trailing spaces and commas

	return &result, nil
}

func (t *Tokenizer) strip(text string, toStrip []string) string {
	for _, strip := range toStrip {
		text = strings.Replace(text, strip, "", -1)
	}
	return text
}
