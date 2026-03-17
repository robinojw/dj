package appserver

import (
	"encoding/json"
	"testing"
)

func TestSubmissionMarshal(t *testing.T) {
	op := UserTurnOp{
		Type:           OpUserTurn,
		Items:          []UserInput{NewTextInput("hello")},
		Cwd:            "/tmp",
		ApprovalPolicy: "never",
		SandboxPolicy:  SandboxPolicyReadOnly(),
		Model:          "o4-mini",
	}
	opData, _ := json.Marshal(op)

	sub := &Submission{
		ID: "sub-1",
		Op: opData,
	}
	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["id"] != "sub-1" {
		t.Errorf("expected id sub-1, got %v", parsed["id"])
	}
}

func TestEventUnmarshal(t *testing.T) {
	raw := `{"id":"","msg":{"type":"session_configured","session_id":"sess-1","model":"gpt-4o"}}`
	var event Event
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.ID != "" {
		t.Errorf("expected empty id, got %s", event.ID)
	}

	var header EventHeader
	if err := json.Unmarshal(event.Msg, &header); err != nil {
		t.Fatal(err)
	}
	if header.Type != EventSessionConfigured {
		t.Errorf("expected session_configured, got %s", header.Type)
	}
}

func TestEventHeaderExtraction(t *testing.T) {
	raw := `{"type":"agent_message_delta","delta":"hello"}`
	var header EventHeader
	if err := json.Unmarshal([]byte(raw), &header); err != nil {
		t.Fatal(err)
	}
	if header.Type != EventAgentMessageDelta {
		t.Errorf("expected agent_message_delta, got %s", header.Type)
	}
}

func TestRPCErrorMessage(t *testing.T) {
	err := &RPCError{Code: -1, Message: "something broke"}
	if err.Error() != "something broke" {
		t.Errorf("expected 'something broke', got %s", err.Error())
	}
}
