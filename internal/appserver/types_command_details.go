package appserver

// CommandDetails holds the command and working directory for approval requests.
type CommandDetails struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}
