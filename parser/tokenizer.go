package parser

import (
	"strings"

	"gopkg.in/jdkato/prose.v2"
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

	result := TokenResult{Tags: []string{}}
	result.Raw = text

	for _, tok := range doc.Tokens() {
		if strings.HasPrefix(tok.Text, "@") || strings.HasPrefix(tok.Text, "#") {
			result.Tags = append(result.Tags, tok.Text)
		}
	}

	return &result, nil
}
