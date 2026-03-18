package appserver

import (
	"encoding/json"
	"testing"
)

const unmarshalFailFormat = "unmarshal: %v"

func TestParseNotification(test *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"thread/started","params":{"thread_id":"t-1"}}`
	var message JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		test.Fatalf(unmarshalFailFormat, err)
	}
	if message.Method != "thread/started" {
		test.Errorf("expected thread/started, got %s", message.Method)
	}
	if message.IsRequest() {
		test.Error("notification should not be a request")
	}
}

func TestParseRequest(test *testing.T) {
	raw := `{"jsonrpc":"2.0","id":"req-1","method":"item/commandExecution/requestApproval","params":{"command":"ls"}}`
	var message JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		test.Fatalf(unmarshalFailFormat, err)
	}
	if message.ID != "req-1" {
		test.Errorf("expected req-1, got %s", message.ID)
	}
	if !message.IsRequest() {
		test.Error("should be a request")
	}
}

func TestParseResponse(test *testing.T) {
	raw := `{"jsonrpc":"2.0","id":"dj-1","result":{"ok":true}}`
	var message JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		test.Fatalf(unmarshalFailFormat, err)
	}
	if !message.IsResponse() {
		test.Error("should be a response")
	}
}
