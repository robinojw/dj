package appserver

import (
	"encoding/json"
	"testing"
)

func TestUserTurnOpMarshal(t *testing.T) {
	op := UserTurnOp{
		Type:           OpUserTurn,
		Items:          []UserInput{NewTextInput("hello")},
		Cwd:            "/home/user",
		ApprovalPolicy: "on-request",
		SandboxPolicy:  SandboxPolicyReadOnly(),
		Model:          "gpt-4o",
	}
	data, err := json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["type"] != "user_turn" {
		t.Errorf("expected user_turn, got %v", parsed["type"])
	}
	if parsed["model"] != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %v", parsed["model"])
	}
}

func TestNewTextInput(t *testing.T) {
	input := NewTextInput("hello world")
	if input.Type != "text" {
		t.Errorf("expected text, got %s", input.Type)
	}
	if input.Text != "hello world" {
		t.Errorf("expected hello world, got %s", input.Text)
	}
	if input.TextElements == nil {
		t.Error("expected non-nil text_elements")
	}
}

func TestSandboxPolicyReadOnly(t *testing.T) {
	policy := SandboxPolicyReadOnly()
	var parsed map[string]any
	json.Unmarshal(policy, &parsed)
	if parsed["type"] != "read-only" {
		t.Errorf("expected read-only, got %v", parsed["type"])
	}
}

func TestSandboxPolicyWorkspaceWrite(t *testing.T) {
	policy := SandboxPolicyWorkspaceWrite([]string{"/home/user/project"})
	var parsed map[string]any
	json.Unmarshal(policy, &parsed)
	if parsed["type"] != "workspace-write" {
		t.Errorf("expected workspace-write, got %v", parsed["type"])
	}
	roots := parsed["writable_roots"].([]any)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
}

func TestInterruptOpMarshal(t *testing.T) {
	op := InterruptOp{Type: OpInterrupt}
	data, _ := json.Marshal(op)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["type"] != "interrupt" {
		t.Errorf("expected interrupt, got %v", parsed["type"])
	}
}
