package appserver

const (
	AgentStatusPendingInit = "pending_init"
	AgentStatusRunning     = "running"
	AgentStatusInterrupted = "interrupted"
	AgentStatusCompleted   = "completed"
	AgentStatusErrored     = "errored"
	AgentStatusShutdown    = "shutdown"
)

type collabBase struct {
	CallID         string `json:"call_id"`
	SenderThreadID string `json:"sender_thread_id"`
}

// CollabSpawnBeginEvent is the params payload for collab/agentSpawn/begin.
type CollabSpawnBeginEvent struct {
	collabBase
	Prompt string `json:"prompt,omitempty"`
	Model  string `json:"model,omitempty"`
}

// CollabSpawnEndEvent is the params payload for collab/agentSpawn/end.
type CollabSpawnEndEvent struct {
	collabBase
	NewThreadID      string `json:"new_thread_id"`
	NewAgentNickname string `json:"new_agent_nickname,omitempty"`
	NewAgentRole     string `json:"new_agent_role,omitempty"`
	Status           string `json:"status"`
}
