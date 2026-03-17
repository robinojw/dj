package appserver

const (
	ThreadStatusActive    = "active"
	ThreadStatusIdle      = "idle"
	ThreadStatusCompleted = "completed"
	ThreadStatusError     = "error"
)

type ThreadStartParams struct {
	Model string `json:"model,omitempty"`
}

type ThreadStartResult struct {
	Thread ThreadInfo `json:"thread"`
}

type ThreadInfo struct {
	ID string `json:"id"`
}

type ThreadArchiveParams struct {
	ThreadID string `json:"threadId"`
}

type ThreadListResult struct {
	Threads []ThreadSummary `json:"threads"`
}

type ThreadSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

type TurnStartParams struct {
	ThreadID string `json:"threadId"`
	Content  string `json:"content"`
}

type TurnStartResult struct {
	TurnID string `json:"turnId"`
}
