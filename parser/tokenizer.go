package parser

import (
	"fmt"
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

	var capturedBrackets strings.Builder
	shouldCapture := false
	for _, tok := range doc.Tokens() {
		if strings.HasPrefix(tok.Text, "@") || strings.HasPrefix(tok.Text, "#") {
			tagMap[tok.Text] = true
		} else if tok.Text == "[" {
			shouldCapture = true
		} else if tok.Text == "]" {
			words := strings.Split(capturedBrackets.String(), ",")
			for _, word := range words {
				trimmed := strings.Trim(word, " ")
				if trimmed == "" {
					continue
				}
				if !strings.HasPrefix(trimmed, "@") && !strings.HasPrefix(trimmed, "#") {
					tagMap[fmt.Sprintf("#%s", trimmed)] = true
				}
			}
			shouldCapture = false
			capturedBrackets.Reset()
		} else if shouldCapture {
			capturedBrackets.WriteString(tok.Text)
		}
	}

	for key, _ := range tagMap {
		result.Tags = append(result.Tags, key)
	}

	return &result, nil
}
