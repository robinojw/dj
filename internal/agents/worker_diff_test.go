package agents

import "testing"

func TestIsEditTool(t *testing.T) {
	tests := []struct {
		name string
		tool string
		want bool
	}{
		{"edit_file is edit tool", "edit_file", true},
		{"write_file is edit tool", "write_file", true},
		{"delete_file is edit tool", "delete_file", true},
		{"read_file is not edit tool", "read_file", false},
		{"bash is not edit tool", "bash", false},
		{"empty string is not edit tool", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEditTool(tt.tool); got != tt.want {
				t.Errorf("isEditTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}
