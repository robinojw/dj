package modes

import "testing"

func TestGate_Evaluate_DenyList(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{"bash(rm -rf*)"})

	// Deny list always wins
	if got := gate.Evaluate("bash(rm -rf /)", nil); got != GateDeny {
		t.Errorf("Expected GateDeny for deny list match, got %v", got)
	}
}

func TestGate_Evaluate_AllowList(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{"write_file"}, []string{})

	// Allow list passes even for write tools in Confirm
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow for allow list match, got %v", got)
	}
}

func TestGate_Evaluate_ConfirmMode(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	tests := []struct {
		tool string
		want GateDecision
	}{
		{"read_file", GateAllow},
		{"write_file", GateAskUser},
		{"bash", GateAskUser},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := gate.Evaluate(tt.tool, nil); got != tt.want {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestGate_Evaluate_PlanMode(t *testing.T) {
	gate := NewGate(ModePlan, []string{}, []string{})

	tests := []struct {
		tool string
		want GateDecision
	}{
		{"read_file", GateAllow},
		{"write_file", GateDeny},
		{"bash", GateDeny},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := gate.Evaluate(tt.tool, nil); got != tt.want {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestGate_Evaluate_TurboMode(t *testing.T) {
	gate := NewGate(ModeTurbo, []string{}, []string{})

	// Turbo allows everything (except deny list)
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow in Turbo mode, got %v", got)
	}
	if got := gate.Evaluate("bash", nil); got != GateAllow {
		t.Errorf("Expected GateAllow in Turbo mode, got %v", got)
	}
}

func TestGate_Evaluate_WithRegistry(t *testing.T) {
	// A custom MCP tool that the static map doesn't know about
	// but the registry says is read-only
	registry := &mockClassifier{readOnly: true, known: true}
	gate := NewGateWithRegistry(ModeConfirm, []string{}, []string{}, registry)

	// Should classify as ToolRead via registry, so auto-allow in Confirm
	if got := gate.Evaluate("custom_mcp_reader", nil); got != GateAllow {
		t.Errorf("Expected GateAllow for registry-known read tool, got %v", got)
	}
}

func TestGate_Evaluate_WithRegistry_UnknownFallsBack(t *testing.T) {
	registry := &mockClassifier{known: false}
	gate := NewGateWithRegistry(ModeConfirm, []string{}, []string{}, registry)

	// Unknown to registry, falls back to static map: bash → ToolExec → AskUser
	if got := gate.Evaluate("bash", nil); got != GateAskUser {
		t.Errorf("Expected GateAskUser for fallback, got %v", got)
	}
}

type mockClassifier struct {
	readOnly     bool
	destructive  bool
	mutatesFiles bool
	known        bool
}

func (m *mockClassifier) ToolAnnotations(name string) (bool, bool, bool, bool) {
	if !m.known {
		return false, false, false, false
	}
	return m.readOnly, m.destructive, m.mutatesFiles, true
}

func TestGate_AllowForSession(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	// Initially asks
	if got := gate.Evaluate("write_file", nil); got != GateAskUser {
		t.Errorf("Expected GateAskUser before session allow, got %v", got)
	}

	// Add to session allow list
	gate.AllowForSession("write_file")

	// Now allows
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow after session allow, got %v", got)
	}
}
