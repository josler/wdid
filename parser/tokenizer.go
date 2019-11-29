package parser

import (
	"strings"

	prose "gopkg.in/jdkato/prose.v2"
)

type TokenResult struct {
	Tags []string
	Raw  string
}

type Tokenizer struct {
}

func (t *Tokenizer) Tokenize(text string) (*TokenResult, error) {
	doc, err := prose.NewDocument(text, prose.WithExtraction(false), prose.WithSegmentation(false), prose.WithTagging(false))
	if err != nil {
		return nil, err
	}

	result := TokenResult{Tags: []string{}, Raw: text}
	tagMap := map[string]bool{}

	for _, tok := range doc.Tokens() {
		if t.isTagPrefix(tok.Text) {
			tagMap[tok.Text] = true
		}
	}

	for key := range tagMap {
		result.Tags = append(result.Tags, key)
	}

	return &result, nil
}

func (t *Tokenizer) isTagPrefix(text string) bool {
	hasAtPrefix := strings.HasPrefix(text, "@") && !strings.HasPrefix(text, "@@")
	hasHashPrefix := strings.HasPrefix(text, "#") && !strings.HasPrefix(text, "##")
	return hasAtPrefix || hasHashPrefix
}
