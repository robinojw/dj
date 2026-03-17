package appserver

import (
	"encoding/json"
	"testing"
)

func TestCommandExecParamsMarshal(t *testing.T) {
	params := CommandExecParams{
		ThreadID: "t-1",
		Command:  "go test ./...",
		TTY:      true,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["tty"] != true {
		t.Errorf("expected tty true, got %v", parsed["tty"])
	}
}

func TestCommandExecResultUnmarshal(t *testing.T) {
	raw := `{"execId":"e-abc123"}`
	var result CommandExecResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.ExecID != "e-abc123" {
		t.Errorf("expected e-abc123, got %s", result.ExecID)
	}
}

func TestConfirmExecParamsUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","command":"rm -rf /tmp/test"}`
	var params ConfirmExecParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", params.ThreadID)
	}
	if params.Command != "rm -rf /tmp/test" {
		t.Errorf("expected command, got %s", params.Command)
	}
}
