package tui

import "time"

type chatMsgKind int

const (
	chatMsgUser      chatMsgKind = iota
	chatMsgAgent
	chatMsgToolCall
	chatMsgToolResult
	chatMsgDiff
	chatMsgError
	chatMsgSystem
)

type chatMsg struct {
	Kind      chatMsgKind
	Content   string
	ToolName  string
	FilePath  string
	DiffLines []string
	Timestamp time.Time
	WorkerID  string
}
