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
	Value    string // the parsed value (path, URL, symbol name, ref, test name)
	Raw      string // the full raw text matched including @
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
	// File mention is the fallback — any @ followed by a path-like string
}

// mentionRegex matches @-prefixed tokens, avoiding email addresses.
// It requires @ to be at the start of the string or preceded by whitespace.
var mentionRegex = regexp.MustCompile(`(?:^|\s)(@(?:fn:|git:|test:|https?://|)[^\s,;]+)`)

// Parse extracts all @mentions from the input string.
func Parse(input string) []Mention {
	var mentions []Mention

	matches := mentionRegex.FindAllStringSubmatchIndex(input, -1)
	for _, match := range matches {
		// match[2] and match[3] are the submatch (group 1) indices
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

	// Remove from right to left to preserve indices
	result := input
	for i := len(mentions) - 1; i >= 0; i-- {
		m := mentions[i]
		result = result[:m.StartIdx] + result[m.EndIdx:]
	}
	return result
}

func classify(raw string) *Mention {
	// Strip leading @
	body := raw[1:]

	// Try typed prefixes first
	for _, p := range parsers {
		prefix := p.prefix[1:] // strip the @ since we already removed it
		if strings.HasPrefix(body, prefix) {
			value := body[len(prefix):]
			return &Mention{
				Type:  p.typ,
				Value: p.extract(value),
				Raw:   raw,
			}
		}
	}

	// Fallback: file mention if it looks like a path
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
	// Must contain a / or a file extension to be a path
	return strings.Contains(s, "/") || strings.Contains(s, ".")
}
