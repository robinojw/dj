package screens

import (
	"reflect"
	"testing"
)

func TestParseDiffLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "single line",
			input: "diff --git a/file.go b/file.go",
			want:  []string{"diff --git a/file.go b/file.go"},
		},
		{
			name:  "multiple lines",
			input: "line1\nline2\nline3",
			want:  []string{"line1", "line2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDiffLines(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDiffLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateDiffStats(t *testing.T) {
	m := &ChatModel{}

	tests := []struct {
		name  string
		lines []string
		want  diffStats
	}{
		{
			name:  "empty diff",
			lines: []string{},
			want:  diffStats{additions: 0, deletions: 0},
		},
		{
			name: "additions only",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"+++ b/file.go",
				"+added line 1",
				"+added line 2",
			},
			want: diffStats{additions: 2, deletions: 0},
		},
		{
			name: "deletions only",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"--- a/file.go",
				"-removed line 1",
				"-removed line 2",
				"-removed line 3",
			},
			want: diffStats{additions: 0, deletions: 3},
		},
		{
			name: "mixed changes",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"--- a/file.go",
				"+++ b/file.go",
				"-removed",
				"+added",
				"+another add",
			},
			want: diffStats{additions: 2, deletions: 1},
		},
		{
			name: "ignores file markers",
			lines: []string{
				"--- a/file.go",
				"+++ b/file.go",
			},
			want: diffStats{additions: 0, deletions: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.calculateDiffStats(tt.lines)
			if got != tt.want {
				t.Errorf("calculateDiffStats() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
