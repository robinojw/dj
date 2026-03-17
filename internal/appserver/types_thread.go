package appserver

const (
	ThreadStatusActive    = "active"
	ThreadStatusIdle      = "idle"
	ThreadStatusCompleted = "completed"
	ThreadStatusError     = "error"
)

type ThreadCreateParams struct {
	Instructions string `json:"instructions"`
}

type ThreadCreateResult struct {
	ThreadID string `json:"threadId"`
}

type ThreadDeleteParams struct {
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

type ThreadSendMessageParams struct {
	ThreadID string `json:"threadId"`
	Content  string `json:"content"`
}

type ThreadSendMessageResult struct {
	MessageID string `json:"messageId"`
}
