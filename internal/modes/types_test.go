package modes

import "testing"

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "Confirm"},
		{ModePlan, "Plan"},
		{ModeTurbo, "Turbo"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyTool(t *testing.T) {
	tests := []struct {
		tool string
		want ToolClass
	}{
		{"read_file", ToolRead},
		{"list_dir", ToolRead},
		{"search_code", ToolRead},
		{"write_file", ToolWrite},
		{"create_file", ToolWrite},
		{"delete_file", ToolWrite},
		{"bash", ToolExec},
		{"run_script", ToolExec},
		{"run_tests", ToolExec},
		{"web_fetch", ToolNetwork},
		{"http_request", ToolNetwork},
		{"unknown_tool", ToolWrite}, // default to conservative
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := ClassifyTool(tt.tool); got != tt.want {
				t.Errorf("ClassifyTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestExecutionMode_StatusLabel(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "⏸ CONFIRM"},
		{ModePlan, "◎ PLAN"},
		{ModeTurbo, "⚡ TURBO"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if got := tt.mode.StatusLabel(); got != tt.want {
				t.Errorf("StatusLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
