package appserver

type ThreadStatusChanged struct {
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
	Title    string `json:"title"`
}

type ItemStarted struct {
	ThreadID string `json:"threadId"`
	ItemID   string `json:"itemId"`
	Role     string `json:"role"`
	Type     string `json:"type"`
}

type ItemCompleted struct {
	ThreadID string `json:"threadId"`
	ItemID   string `json:"itemId"`
	Content  string `json:"content"`
}

type ItemMessageDelta struct {
	ThreadID string `json:"threadId"`
	ItemID   string `json:"itemId"`
	Delta    string `json:"delta"`
}

type TurnStarted struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type TurnCompleted struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
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
