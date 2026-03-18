package appserver

import (
	"encoding/json"
	"testing"
)

func TestSessionConfiguredUnmarshal(t *testing.T) {
	raw := `{"type":"session_configured","session_id":"s-123","model":"o4-mini","reasoning_effort":"medium","history_log_id":0}`
	var cfg SessionConfigured
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.SessionID != "s-123" {
		t.Errorf("expected s-123, got %s", cfg.SessionID)
	}
	if cfg.Model != "o4-mini" {
		t.Errorf("expected o4-mini, got %s", cfg.Model)
	}
}

func TestTaskStartedUnmarshal(t *testing.T) {
	raw := `{"type":"task_started","model_context_window":200000}`
	var started TaskStarted
	if err := json.Unmarshal([]byte(raw), &started); err != nil {
		t.Fatal(err)
	}
	if started.ModelContextWindow != 200000 {
		t.Errorf("expected 200000, got %d", started.ModelContextWindow)
	}
}

func TestTaskCompleteUnmarshal(t *testing.T) {
	raw := `{"type":"task_complete","last_agent_message":"Hello"}`
	var complete TaskComplete
	if err := json.Unmarshal([]byte(raw), &complete); err != nil {
		t.Fatal(err)
	}
	if complete.LastAgentMessage != "Hello" {
		t.Errorf("expected Hello, got %s", complete.LastAgentMessage)
	}
}

func TestAgentDeltaUnmarshal(t *testing.T) {
	raw := `{"type":"agent_message_delta","delta":"Howdy"}`
	var delta AgentDelta
	if err := json.Unmarshal([]byte(raw), &delta); err != nil {
		t.Fatal(err)
	}
	if delta.Delta != "Howdy" {
		t.Errorf("expected Howdy, got %s", delta.Delta)
	}
}

func TestAgentMessageUnmarshal(t *testing.T) {
	raw := `{"type":"agent_message","message":"Hello world"}`
	var msg AgentMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Message != "Hello world" {
		t.Errorf("expected Hello world, got %s", msg.Message)
	}
}

func TestUserInputOpMarshal(t *testing.T) {
	op := UserInputOp{
		Type: OpUserInput,
		Items: []InputItem{
			{Type: "text", Text: "Say hello"},
		},
	}
	data, err := json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["type"] != "user_input" {
		t.Errorf("expected user_input, got %v", parsed["type"])
	}
}

func TestExecCommandRequestUnmarshal(t *testing.T) {
	raw := `{"type":"exec_command_request","command":"ls -la","cwd":"/tmp"}`
	var req ExecCommandRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Command != "ls -la" {
		t.Errorf("expected ls -la, got %s", req.Command)
	}
	if req.Cwd != "/tmp" {
		t.Errorf("expected /tmp, got %s", req.Cwd)
	}
}
