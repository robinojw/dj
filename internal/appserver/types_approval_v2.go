package appserver

// CommandApprovalRequest is the v2 params payload for command execution approval.
type CommandApprovalRequest struct {
	threadScoped
	Command CommandDetails `json:"command"`
}

// FileChangeApprovalRequest is the v2 params payload for file change approval.
type FileChangeApprovalRequest struct {
	threadScoped
	Patch string `json:"patch"`
}
