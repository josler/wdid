package parser

import (
	"regexp"
)

type TokenResult struct {
	Tags        []string
	Connections []string
	Raw         string
}

type Tokenizer struct {
}

func (t *Tokenizer) Tokenize(text string) (*TokenResult, error) {
	return &TokenResult{
		Tags:        t.getTags(text),
		Connections: t.getConnections(text),
		Raw:         text,
	}, nil
}

func (t *Tokenizer) getConnections(text string) []string {
	re := regexp.MustCompile(`\[\[([^\s\[\]]+)\]\]`)
	found := re.FindAllStringSubmatch(text, -1)
	connections := []string{}
	for _, f := range found {
		connections = append(connections, f[1])
	}
	return connections
}

func (t *Tokenizer) getTags(text string) []string {
	tagExp := regexp.MustCompile(`(^|[^\w#])(#[\w]+)`)
	tags := tagExp.FindAllStringSubmatch(text, -1)

	mentionExp := regexp.MustCompile(`(^|[^\w@])(@[\w]+)`)
	mentions := mentionExp.FindAllStringSubmatch(text, -1)

	tagMentionsMap := map[string]bool{}
	result := []string{}

	for _, tag := range tags {
		if _, ok := tagMentionsMap[tag[2]]; !ok {
			tagMentionsMap[tag[2]] = true
			result = append(result, tag[2])
		}
	}
	for _, mention := range mentions {
		if _, ok := tagMentionsMap[mention[2]]; !ok {
			tagMentionsMap[mention[2]] = true
			result = append(result, mention[2])
		}
	}

	return result
}
