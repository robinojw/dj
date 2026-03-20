package tui

import "github.com/robinojw/dj/internal/appserver"

type SpawnRequestMsg struct {
	SourceAgentID string
	Persona       string
	Task          string
}

type AgentMessageMsg struct {
	SourceAgentID string
	TargetAgentID string
	Content       string
}

type AgentCompleteMsg struct {
	AgentID string
	Content string
}

type PoolEventMsg struct {
	AgentID string
	Message appserver.JSONRPCMessage
}
