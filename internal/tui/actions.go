package tui

type ForkThreadMsg struct {
	ParentID     string
	Instructions string
}

type DeleteThreadMsg struct {
	ThreadID string
}

type RenameThreadMsg struct {
	ThreadID string
	NewTitle string
}
