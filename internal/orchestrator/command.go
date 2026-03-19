package orchestrator

import (
	"encoding/json"
	"strings"
)

const (
	fenceOpen  = "```dj-command\n"
	fenceClose = "\n```"
)

type DJCommand struct {
	Action  string `json:"action"`
	Persona string `json:"persona,omitempty"`
	Task    string `json:"task,omitempty"`
	Target  string `json:"target,omitempty"`
	Content string `json:"content,omitempty"`
}

type CommandParser struct {
	buffer      strings.Builder
	commands    []DJCommand
	cleanedText strings.Builder
}

func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

func (parser *CommandParser) Feed(delta string) {
	parser.buffer.WriteString(delta)
}

func (parser *CommandParser) Flush() []DJCommand {
	parser.commands = nil
	parser.cleanedText.Reset()

	text := parser.buffer.String()
	parser.buffer.Reset()

	for {
		openIndex := strings.Index(text, fenceOpen)
		if openIndex == -1 {
			parser.cleanedText.WriteString(text)
			break
		}

		parser.cleanedText.WriteString(text[:openIndex])
		rest := text[openIndex+len(fenceOpen):]

		closeIndex := strings.Index(rest, fenceClose)
		if closeIndex == -1 {
			parser.buffer.WriteString(text[openIndex:])
			break
		}

		jsonBlock := strings.TrimSpace(rest[:closeIndex])
		var command DJCommand
		if err := json.Unmarshal([]byte(jsonBlock), &command); err == nil {
			parser.commands = append(parser.commands, command)
		}

		remaining := rest[closeIndex+len(fenceClose):]
		text = strings.TrimPrefix(remaining, "\n")
	}

	return parser.commands
}

func (parser *CommandParser) CleanedText() string {
	return parser.cleanedText.String()
}
