package mentions

import (
	"regexp"
	"strings"
)

// MentionType categorizes the kind of @mention.
type MentionType string

const (
	MentionFile     MentionType = "file"
	MentionURL      MentionType = "url"
	MentionFunction MentionType = "function"
	MentionGit      MentionType = "git"
	MentionTest     MentionType = "test"
)

// Mention represents a parsed @mention from user input.
type Mention struct {
	Type     MentionType
	Value    string
	Raw      string
	StartIdx int
	EndIdx   int
}

// ResolvedMention is a mention with its fetched content.
type ResolvedMention struct {
	Mention
	Content string
	Error   error
}

// parser is a typed prefix handler.
type parser struct {
	prefix  string
	typ     MentionType
	extract func(string) string
}

var parsers = []parser{
	{prefix: "@fn:", typ: MentionFunction, extract: func(s string) string { return s }},
	{prefix: "@git:", typ: MentionGit, extract: func(s string) string { return s }},
	{prefix: "@test:", typ: MentionTest, extract: func(s string) string { return s }},
	{prefix: "@https://", typ: MentionURL, extract: func(s string) string { return "https://" + s }},
	{prefix: "@http://", typ: MentionURL, extract: func(s string) string { return "http://" + s }},
}

// mentionRegex matches @-prefixed tokens, avoiding email addresses.
// It requires @ to be at the start of the string or preceded by whitespace.
var mentionRegex = regexp.MustCompile(`(?:^|\s)(@(?:fn:|git:|test:|https?://|)[^\s,;]+)`)

// Parse extracts all @mentions from the input string.
func Parse(input string) []Mention {
	var mentions []Mention

	matches := mentionRegex.FindAllStringSubmatchIndex(input, -1)
	for _, match := range matches {
		raw := input[match[2]:match[3]]

		m := classify(raw)
		if m != nil {
			m.StartIdx = match[2]
			m.EndIdx = match[3]
			mentions = append(mentions, *m)
		}
	}

	return mentions
}

// StripMentions removes all @mention tokens from the input.
func StripMentions(input string) string {
	mentions := Parse(input)
	if len(mentions) == 0 {
		return input
	}

	result := input
	for i := len(mentions) - 1; i >= 0; i-- {
		m := mentions[i]
		result = result[:m.StartIdx] + result[m.EndIdx:]
	}
	return result
}

func classify(raw string) *Mention {
	body := raw[1:]

	for _, p := range parsers {
		prefix := p.prefix[1:]
		if strings.HasPrefix(body, prefix) {
			value := body[len(prefix):]
			return &Mention{
				Type:  p.typ,
				Value: p.extract(value),
				Raw:   raw,
			}
		}
	}

	if isPathLike(body) {
		return &Mention{
			Type:  MentionFile,
			Value: body,
			Raw:   raw,
		}
	}

	return nil
}

func isPathLike(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, ".")
}
