package appserver

type CommandExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
	TTY      bool   `json:"tty"`
}

type CommandExecResult struct {
	ExecID string `json:"execId"`
}

type ConfirmExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
}

type ConfirmExecResult struct {
	Approved bool `json:"approved"`
}
