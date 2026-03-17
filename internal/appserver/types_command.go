package appserver

// CommandExecParams is used when the TUI wants to execute a command.
type CommandExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
	TTY      bool   `json:"tty"`
}

// CommandExecResult is the result of a command execution.
type CommandExecResult struct {
	ExecID string `json:"execId"`
}

// ConfirmExecParams is used for exec approval responses.
type ConfirmExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
}

// ConfirmExecResult is the approval response.
type ConfirmExecResult struct {
	Approved bool `json:"approved"`
}
