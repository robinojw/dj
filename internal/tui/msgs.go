package tui

type ThreadStatusMsg struct {
	ThreadID string
	Status   string
	Title    string
}

type ThreadMessageMsg struct {
	ThreadID  string
	MessageID string
	Role      string
	Content   string
}

type ThreadDeltaMsg struct {
	ThreadID  string
	MessageID string
	Delta     string
}

type CommandOutputMsg struct {
	ThreadID string
	ExecID   string
	Data     string
}

type CommandFinishedMsg struct {
	ThreadID string
	ExecID   string
	ExitCode int
}

type ThreadCreatedMsg struct {
	ThreadID string
	Title    string
}

type ThreadDeletedMsg struct {
	ThreadID string
}

type AppServerErrorMsg struct {
	Err error
}

func (msg AppServerErrorMsg) Error() string {
	return msg.Err.Error()
}
