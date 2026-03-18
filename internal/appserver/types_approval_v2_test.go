package appserver

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalCommandApproval(test *testing.T) {
	raw := `{"thread_id":"t-1","command":{"command":"ls -la","cwd":"/tmp"}}`
	var request CommandApprovalRequest
	if err := json.Unmarshal([]byte(raw), &request); err != nil {
		test.Fatalf("approval unmarshal: %v", err)
	}
	if request.ThreadID != "t-1" {
		test.Errorf("expected t-1, got %s", request.ThreadID)
	}
	if request.Command.Command != "ls -la" {
		test.Errorf("expected ls -la, got %s", request.Command.Command)
	}
}
