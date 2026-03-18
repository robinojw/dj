package appserver

type SessionConfigured struct {
	SessionID       string `json:"session_id"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
	HistoryLogID    int64  `json:"history_log_id"`
	RolloutPath     string `json:"rollout_path"`
}

type TaskStarted struct {
	ModelContextWindow int `json:"model_context_window"`
}

type TaskComplete struct {
	LastAgentMessage string `json:"last_agent_message"`
}

type AgentMessage struct {
	Message string `json:"message"`
}

type AgentDelta struct {
	Delta string `json:"delta"`
}

type UserInputOp struct {
	Type  string      `json:"type"`
	Items []InputItem `json:"items"`
}

type InputItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ExecCommandRequest struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}

type PatchApplyRequest struct {
	Patch string `json:"patch"`
}
