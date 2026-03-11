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

func TestClassifyToolWithRegistry(t *testing.T) {
	tests := []struct {
		name string
		tool string
		ann  *mockAnnotations
		want ToolClass
	}{
		{
			name: "registry says read-only",
			tool: "custom_reader",
			ann:  &mockAnnotations{readOnly: true, known: true},
			want: ToolRead,
		},
		{
			name: "registry says destructive",
			tool: "custom_writer",
			ann:  &mockAnnotations{destructive: true, known: true},
			want: ToolWrite,
		},
		{
			name: "registry says mutates files",
			tool: "custom_editor",
			ann:  &mockAnnotations{mutatesFiles: true, known: true},
			want: ToolWrite,
		},
		{
			name: "nil registry falls back to static map",
			tool: "read_file",
			ann:  nil,
			want: ToolRead,
		},
		{
			name: "unknown tool with nil registry defaults to ToolWrite",
			tool: "mystery",
			ann:  nil,
			want: ToolWrite,
		},
		{
			name: "registry has no annotations, falls back",
			tool: "bash",
			ann:  &mockAnnotations{known: false},
			want: ToolExec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tc ToolClassifier
			if tt.ann != nil {
				tc = tt.ann
			}
			if got := ClassifyToolWithRegistry(tt.tool, tc); got != tt.want {
				t.Errorf("ClassifyToolWithRegistry(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

type mockAnnotations struct {
	readOnly     bool
	destructive  bool
	mutatesFiles bool
	known        bool
}

func (m *mockAnnotations) ToolAnnotations(name string) (readOnly, destructive, mutatesFiles bool, ok bool) {
	if !m.known {
		return false, false, false, false
	}
	return m.readOnly, m.destructive, m.mutatesFiles, true
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
