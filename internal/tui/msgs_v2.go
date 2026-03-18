package tui

// ThreadStartedMsg is emitted when a new thread is created via v2 protocol.
type ThreadStartedMsg struct {
	ThreadID      string
	Status        string
	SourceType    string
	ParentID      string
	Depth         int
	AgentNickname string
	AgentRole     string
}

// ThreadStatusChangedMsg is emitted when a thread's status changes.
type ThreadStatusChangedMsg struct {
	ThreadID string
	Status   string
}

// TurnStartedMsg is emitted when a turn begins in a thread.
type TurnStartedMsg struct {
	ThreadID string
	TurnID   string
}

// TurnCompletedMsg is emitted when a turn finishes in a thread.
type TurnCompletedMsg struct {
	ThreadID string
	TurnID   string
}

// V2AgentDeltaMsg is a streaming text delta scoped to a thread.
type V2AgentDeltaMsg struct {
	ThreadID string
	Delta    string
}

// V2ExecApprovalMsg is a v2 command execution approval request.
type V2ExecApprovalMsg struct {
	RequestID string
	ThreadID  string
	Command   string
	Cwd       string
}

// V2FileApprovalMsg is a v2 file change approval request.
type V2FileApprovalMsg struct {
	RequestID string
	ThreadID  string
	Patch     string
}

// CollabSpawnMsg is emitted when a sub-agent is spawned.
type CollabSpawnMsg struct {
	SenderThreadID   string
	NewThreadID      string
	NewAgentNickname string
	NewAgentRole     string
	Status           string
}

// CollabCloseMsg is emitted when a sub-agent is closed.
type CollabCloseMsg struct {
	SenderThreadID   string
	ReceiverThreadID string
	Status           string
}

// CollabStatusUpdateMsg is emitted for general collab status changes.
type CollabStatusUpdateMsg struct {
	ThreadID string
	Status   string
}
