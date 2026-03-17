package appserver

import "encoding/json"

// UserTurnOp is the "user_turn" submission operation.
type UserTurnOp struct {
	Type           string          `json:"type"`
	Items          []UserInput     `json:"items"`
	Cwd            string          `json:"cwd"`
	ApprovalPolicy string          `json:"approval_policy"`
	SandboxPolicy  json.RawMessage `json:"sandbox_policy"`
	Model          string          `json:"model"`
}

// UserInput is a content item in a user turn.
type UserInput struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	TextElements []any  `json:"text_elements,omitempty"`
}

// NewTextInput creates a text user input item.
func NewTextInput(text string) UserInput {
	return UserInput{
		Type:         "text",
		Text:         text,
		TextElements: []any{},
	}
}

// SandboxPolicyReadOnly creates a read-only sandbox policy.
func SandboxPolicyReadOnly() json.RawMessage {
	return json.RawMessage(`{"type":"read-only","network_access":false}`)
}

// SandboxPolicyWorkspaceWrite creates a workspace-write sandbox policy.
func SandboxPolicyWorkspaceWrite(roots []string) json.RawMessage {
	policy := map[string]any{
		"type":           "workspace-write",
		"writable_roots": roots,
		"network_access": false,
	}
	data, _ := json.Marshal(policy)
	return data
}

// InterruptOp is the "interrupt" submission operation.
type InterruptOp struct {
	Type string `json:"type"`
}

// ShutdownOp is the "shutdown" submission operation.
type ShutdownOp struct {
	Type string `json:"type"`
}

// ExecApprovalOp is the "exec_approval" submission operation.
type ExecApprovalOp struct {
	Type     string `json:"type"`
	Approved bool   `json:"approved"`
}
