package agents

import "testing"

func TestModeConfigPlan(t *testing.T) {
	cfg := Modes[ModePlan]
	if cfg.ReasoningEffort != "high" {
		t.Errorf("Plan mode should use high reasoning, got %s", cfg.ReasoningEffort)
	}
	if len(cfg.AllowedTools) == 0 {
		t.Error("Plan mode should have an explicit allowlist")
	}
	for _, tool := range cfg.AllowedTools {
		if tool == "write_file" || tool == "run_command" {
			t.Errorf("Plan mode should not allow %s", tool)
		}
	}
}

func TestModeConfigBuild(t *testing.T) {
	cfg := Modes[ModeBuild]
	if cfg.ReasoningEffort != "medium" {
		t.Errorf("Build mode should use medium reasoning, got %s", cfg.ReasoningEffort)
	}
	if cfg.AllowedTools != nil {
		t.Error("Build mode should allow all tools (nil allowlist)")
	}
}

func TestFilterToolsByMode(t *testing.T) {
	allTools := []string{"read_file", "write_file", "run_command", "search_code", "list_dir"}
	cfg := Modes[ModePlan]
	filtered := FilterTools(allTools, cfg)

	for _, tool := range filtered {
		if tool == "write_file" || tool == "run_command" {
			t.Errorf("Plan mode filtered list should not contain %s", tool)
		}
	}
	if len(filtered) != 3 {
		t.Errorf("Expected 3 tools in plan mode, got %d", len(filtered))
	}
}
