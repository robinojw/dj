package tui

type ThreadCreatedMsg struct {
	ThreadID string
	Title    string
}

type ThreadDeletedMsg struct {
	ThreadID string
}

type PTYOutputMsg struct {
	ThreadID string
	Exited   bool
}

type FocusPane int

const (
	FocusPaneCanvas  FocusPane = iota
	FocusPaneSession
)

type PinSessionMsg struct {
	ThreadID string
}

type UnpinSessionMsg struct {
	ThreadID string
}

type FocusSessionPaneMsg struct {
	Index int
}

type SwitchPaneFocusMsg struct {
	Pane FocusPane
}

type AppServerErrorMsg struct {
	Err error
}

func (msg AppServerErrorMsg) Error() string {
	return msg.Err.Error()
}
