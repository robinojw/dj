package modes

import (
	"testing"
	"time"
)

func TestPermissionRequest_ResponseFlow(t *testing.T) {
	req := PermissionRequest{
		ID:       "test-123",
		WorkerID: "worker-1",
		Tool:     "write_file",
		Args:     map[string]any{"path": "test.go"},
		RespCh:   make(chan PermissionResp, 1),
	}

	// Simulate user approval
	go func() {
		req.RespCh <- PermissionResp{
			Allowed:     true,
			RememberFor: RememberSession,
		}
	}()

	// Worker blocks waiting for response
	resp := <-req.RespCh

	if !resp.Allowed {
		t.Error("Expected approval")
	}
	if resp.RememberFor != RememberSession {
		t.Errorf("Expected RememberSession, got %v", resp.RememberFor)
	}
}

func TestPermissionRequest_Timeout(t *testing.T) {
	req := PermissionRequest{
		ID:     "test-456",
		RespCh: make(chan PermissionResp, 1),
	}

	// Simulate timeout
	select {
	case <-req.RespCh:
		t.Error("Should not receive response")
	case <-time.After(10 * time.Millisecond):
		// Expected timeout
	}
}
