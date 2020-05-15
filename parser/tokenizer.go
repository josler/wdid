package parser

import (
	"strings"

	prose "gopkg.in/jdkato/prose.v2"
)

type TokenResult struct {
	Tags        []string
	Connections []string
	Raw         string
}

type Tokenizer struct {
}

func (t *Tokenizer) Tokenize(text string) (*TokenResult, error) {
	doc, err := prose.NewDocument(text, prose.WithExtraction(false), prose.WithSegmentation(false), prose.WithTagging(false))
	if err != nil {
		return nil, err
	}

	result := TokenResult{Tags: []string{}, Connections: []string{}, Raw: text}
	tagMap := map[string]bool{}
	startConnectionCounter := 0
	endConnectionCounter := 0

	var capturedBrackets strings.Builder

	for _, tok := range doc.Tokens() {
		if t.isTagPrefix(tok.Text) {
			tagMap[tok.Text] = true
		} else if tok.Text == "[" {
			if startConnectionCounter >= 2 {
				startConnectionCounter = 0
				endConnectionCounter = 0
				capturedBrackets.Reset()
			}
			// add one to starting counter if we see
			startConnectionCounter += 1
			endConnectionCounter = 0
		} else if tok.Text == "]" {
			// ignore a closing bracket and reset everything unless the starting counter
			// is complete (at 2)
			if startConnectionCounter != 2 {
				startConnectionCounter = 0
				endConnectionCounter = 0
				capturedBrackets.Reset()
				continue
			}
			// otherwise, add one to the counter
			endConnectionCounter += 1

			// don't do anything else unless we've 2 end counters (and 2 start counters)
			if endConnectionCounter != 2 {
				continue
			}

			// parse and store
			trimmed := strings.Trim(capturedBrackets.String(), " ")
			if trimmed != "" {
				result.Connections = append(result.Connections, trimmed)
			}

			// reset
			startConnectionCounter = 0
			endConnectionCounter = 0
			capturedBrackets.Reset()
		} else if startConnectionCounter == 2 && endConnectionCounter == 0 {
			// if we're in an "open" state, then capture things
			capturedBrackets.WriteString(tok.Text)
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
