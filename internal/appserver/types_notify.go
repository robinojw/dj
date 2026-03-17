package appserver

// SessionConfigured is the initial event sent by the server on startup.
type SessionConfigured struct {
	Type              string `json:"type"`
	SessionID         string `json:"session_id"`
	Model             string `json:"model"`
	ReasoningEffort   string `json:"reasoning_effort"`
	HistoryLogID      int    `json:"history_log_id"`
	HistoryEntryCount int    `json:"history_entry_count"`
	RolloutPath       string `json:"rollout_path"`
}

// TaskStarted signals the beginning of an agent turn.
type TaskStarted struct {
	Type string `json:"type"`
}

// TaskComplete signals the end of an agent turn.
type TaskComplete struct {
	Type string `json:"type"`
}

// AgentMessage is a complete agent message.
type AgentMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// AgentMessageDelta is a streaming text delta from the agent.
type AgentMessageDelta struct {
	Type  string `json:"type"`
	Delta string `json:"delta"`
}

// ExecCommandBegin signals the start of a command execution.
type ExecCommandBegin struct {
	Type    string `json:"type"`
	ExecID  string `json:"call_id"`
	Command string `json:"command"`
}

// ExecCommandOutputDelta is a chunk of command output.
type ExecCommandOutputDelta struct {
	Type   string `json:"type"`
	ExecID string `json:"call_id"`
	Delta  string `json:"delta"`
}

// ExecCommandEnd signals the end of a command execution.
type ExecCommandEnd struct {
	Type     string `json:"type"`
	ExecID   string `json:"call_id"`
	ExitCode int    `json:"exit_code"`
}

// ExecApprovalRequest asks the client to approve a command.
type ExecApprovalRequest struct {
	Type    string `json:"type"`
	ExecID  string `json:"call_id"`
	Command string `json:"command"`
}

// ServerError is an error event from the server.
type ServerError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
