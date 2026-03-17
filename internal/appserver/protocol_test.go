package appserver

import (
	"encoding/json"
	"testing"
)

func TestRequestMarshal(t *testing.T) {
	req := &Request{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "thread/list",
		Params:  json.RawMessage(`{}`),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %v", parsed["jsonrpc"])
	}
	if parsed["method"] != "thread/list" {
		t.Errorf("expected method thread/list, got %v", parsed["method"])
	}
}

func TestResponseUnmarshal(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"result":{"threads":[]}}`
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID == nil || *resp.ID != 1 {
		t.Errorf("expected id 1, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Error("expected no error")
	}
}

func TestNotificationUnmarshal(t *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"thread/status/changed","params":{"threadId":"t1","status":"active"}}`
	var msg Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Method != "thread/status/changed" {
		t.Errorf("expected thread/status/changed, got %s", msg.Method)
	}
	// Notification: no ID
	if msg.ID != nil {
		t.Error("notification should have no id")
	}
}

func TestErrorResponseUnmarshal(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":2,"error":{"code":-32600,"message":"Invalid Request"}}`
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("expected code -32600, got %d", resp.Error.Code)
	}
}

func intPtr(i int) *int { return &i }
