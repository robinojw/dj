package appserver

import (
	"encoding/json"
	"testing"
)

func TestProtoEventUnmarshal(t *testing.T) {
	raw := `{"id":"","msg":{"type":"session_configured","session_id":"s1","model":"o4-mini"}}`
	var event ProtoEvent
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
	if header.Type != "session_configured" {
		t.Errorf("expected session_configured, got %s", header.Type)
	}
}

func TestProtoEventWithID(t *testing.T) {
	raw := `{"id":"req-1","msg":{"type":"exec_command_request","command":"ls"}}`
	var event ProtoEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event.ID != "req-1" {
		t.Errorf("expected req-1, got %s", event.ID)
	}
}

func TestProtoSubmissionMarshal(t *testing.T) {
	op, _ := json.Marshal(map[string]string{"type": "user_input"})
	sub := &ProtoSubmission{
		ID: "dj-1",
		Op: op,
	}
	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["id"] != "dj-1" {
		t.Errorf("expected dj-1, got %v", parsed["id"])
	}
}
