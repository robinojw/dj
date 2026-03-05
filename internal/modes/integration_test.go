package modes

import (
	"testing"
	"time"
)

func TestIntegration_PermissionFlow(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	// Simulate worker requesting permission
	respCh := make(chan PermissionResp, 1)
	req := PermissionRequest{
		ID:       "int-test-1",
		WorkerID: "worker-1",
		Tool:     "write_file",
		Args:     map[string]any{"path": "test.go"},
		RespCh:   respCh,
	}

	// Simulate worker evaluating
	decision := gate.Evaluate(req.Tool, req.Args)
	if decision != GateAskUser {
		t.Fatalf("Expected GateAskUser, got %v", decision)
	}

	// Simulate TUI approval
	go func() {
		time.Sleep(10 * time.Millisecond)
		respCh <- PermissionResp{
			Allowed:     true,
			RememberFor: RememberSession,
		}
	}()

	// Worker blocks
	resp := <-respCh

	if !resp.Allowed {
		t.Error("Expected approval")
	}
	if resp.RememberFor != RememberSession {
		t.Errorf("Expected RememberSession, got %v", resp.RememberFor)
	}

	// Add to session allow list
	gate.AllowForSession(req.Tool)

	// Second call should auto-allow
	decision = gate.Evaluate(req.Tool, req.Args)
	if decision != GateAllow {
		t.Errorf("Expected GateAllow after session allow, got %v", decision)
	}
}

func TestIntegration_ModeCycle(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		tool string
		want GateDecision
	}{
		{ModeConfirm, "read_file", GateAllow},
		{ModeConfirm, "write_file", GateAskUser},
		{ModePlan, "read_file", GateAllow},
		{ModePlan, "write_file", GateDeny},
		{ModeTurbo, "read_file", GateAllow},
		{ModeTurbo, "write_file", GateAllow},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String()+"_"+tt.tool, func(t *testing.T) {
			gate := NewGate(tt.mode, []string{}, []string{})
			decision := gate.Evaluate(tt.tool, nil)
			if decision != tt.want {
				t.Errorf("Mode %s, tool %s: got %v, want %v",
					tt.mode, tt.tool, decision, tt.want)
			}
		})
	}
}
