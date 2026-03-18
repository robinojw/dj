package appserver

const (
	SourceTypeCLI      = "cli"
	SourceTypeSubAgent = "sub_agent"
	SourceTypeExec     = "exec"
)

const (
	ThreadStatusIdle   = "idle"
	ThreadStatusActive = "active"
)

// SessionSource describes how a thread was created.
type SessionSource struct {
	Type           string `json:"type"`
	ParentThreadID string `json:"parent_thread_id,omitempty"`
	Depth          int    `json:"depth,omitempty"`
	AgentNickname  string `json:"agent_nickname,omitempty"`
	AgentRole      string `json:"agent_role,omitempty"`
}

// Thread represents a v2 thread object within notifications.
type Thread struct {
	ID     string        `json:"id"`
	Status string        `json:"status"`
	Source SessionSource `json:"source"`
}

// ThreadStartedNotification is the params payload for thread/started.
type ThreadStartedNotification struct {
	Thread Thread `json:"thread"`
}
