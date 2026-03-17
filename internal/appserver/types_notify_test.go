package appserver

import (
	"encoding/json"
	"testing"
)

func TestSessionConfiguredUnmarshal(t *testing.T) {
	raw := `{"type":"session_configured","session_id":"sess-1","model":"gpt-4o","reasoning_effort":"high","history_log_id":123,"history_entry_count":5,"rollout_path":"/tmp/rollout.jsonl"}`
	var event SessionConfigured
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.SessionID != "sess-1" {
		t.Errorf("expected sess-1, got %s", event.SessionID)
	}
	if event.Model != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", event.Model)
	}
}

func TestAgentMessageDeltaUnmarshal(t *testing.T) {
	raw := `{"type":"agent_message_delta","delta":"hello world"}`
	var event AgentMessageDelta
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.Delta != "hello world" {
		t.Errorf("expected 'hello world', got %s", event.Delta)
	}
}

func TestExecCommandBeginUnmarshal(t *testing.T) {
	raw := `{"type":"exec_command_begin","call_id":"cmd-1","command":"ls -la"}`
	var event ExecCommandBegin
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.ExecID != "cmd-1" {
		t.Errorf("expected cmd-1, got %s", event.ExecID)
	}
	if event.Command != "ls -la" {
		t.Errorf("expected ls -la, got %s", event.Command)
	}
}

func TestExecCommandOutputDeltaUnmarshal(t *testing.T) {
	raw := `{"type":"exec_command_output_delta","call_id":"cmd-1","delta":"output line\n"}`
	var event ExecCommandOutputDelta
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.Delta != "output line\n" {
		t.Errorf("expected output, got %s", event.Delta)
	}
}

func TestExecCommandEndUnmarshal(t *testing.T) {
	raw := `{"type":"exec_command_end","call_id":"cmd-1","exit_code":0}`
	var event ExecCommandEnd
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", event.ExitCode)
	}
}

func TestServerErrorUnmarshal(t *testing.T) {
	raw := `{"type":"error","message":"something went wrong"}`
	var event ServerError
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.Message != "something went wrong" {
		t.Errorf("expected error message, got %s", event.Message)
	}
}

func TestExecApprovalRequestUnmarshal(t *testing.T) {
	raw := `{"type":"exec_approval_request","call_id":"cmd-1","command":"rm -rf /tmp/test"}`
	var event ExecApprovalRequest
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.Command != "rm -rf /tmp/test" {
		t.Errorf("expected command, got %s", event.Command)
	}
}
