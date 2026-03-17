package appserver

type ThreadStatusChanged struct {
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
	Title    string `json:"title"`
}

type ThreadMessageCreated struct {
	ThreadID  string `json:"threadId"`
	MessageID string `json:"messageId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

type ThreadMessageDelta struct {
	ThreadID  string `json:"threadId"`
	MessageID string `json:"messageId"`
	Delta     string `json:"delta"`
}

type CommandOutput struct {
	ThreadID string `json:"threadId"`
	ExecID   string `json:"execId"`
	Data     string `json:"data"`
}

type CommandFinished struct {
	ThreadID string `json:"threadId"`
	ExecID   string `json:"execId"`
	ExitCode int    `json:"exitCode"`
}
