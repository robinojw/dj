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
