package modes

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"bash(git status*)", "bash(git status)", true},
		{"bash(git status*)", "bash(git status --short)", true},
		{"bash(git status*)", "bash(git diff)", false},
		{"read_file(.env*)", "read_file(.env)", true},
		{"read_file(.env*)", "read_file(.env.local)", true},
		{"read_file(.env*)", "read_file(config.toml)", false},
		{"exact_match", "exact_match", true},
		{"exact_match", "exact_match_not", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"~"+tt.input, func(t *testing.T) {
			if got := MatchGlob(tt.pattern, tt.input); got != tt.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}
