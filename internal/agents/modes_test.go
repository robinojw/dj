package agents

import "testing"

func TestFilterTools(t *testing.T) {
	allTools := []string{"read_file", "write_file", "bash"}

	// Plan mode filters to read-only
	planCfg := Modes[ModePlan]
	filtered := FilterTools(allTools, planCfg)

	if len(filtered) != 1 || filtered[0] != "read_file" {
		t.Errorf("Expected only read_file, got %v", filtered)
	}

	// Confirm/Turbo allow all tools
	confirmCfg := Modes[ModeConfirm]
	filtered = FilterTools(allTools, confirmCfg)

	if len(filtered) != 3 {
		t.Errorf("Expected all 3 tools, got %d", len(filtered))
	}
}
