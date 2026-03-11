package agents

import (
	"os"
	"os/exec"
	"testing"

	"github.com/robinojw/dj/internal/tools"
)

func TestIsDestructiveTool_WithRegistry(t *testing.T) {
	registry := tools.NewDefaultRegistry(t.TempDir())
	w := &Worker{registry: registry}

	tests := []struct {
		name string
		tool string
		want bool
	}{
		{"edit_file is destructive", "edit_file", true},
		{"write_file is destructive", "write_file", true},
		{"str_replace is destructive", "str_replace", true},
		{"read_file is not destructive", "read_file", false},
		{"list_dir is not destructive", "list_dir", false},
		{"run_tests is not destructive", "run_tests", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w.isDestructiveTool(tt.tool); got != tt.want {
				t.Errorf("isDestructiveTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestIsDestructiveTool_Fallback(t *testing.T) {
	// Worker with nil registry uses hardcoded fallback
	w := &Worker{}

	tests := []struct {
		name string
		tool string
		want bool
	}{
		{"edit_file is destructive", "edit_file", true},
		{"write_file is destructive", "write_file", true},
		{"delete_file is destructive", "delete_file", true},
		{"read_file is not destructive", "read_file", false},
		{"bash is not destructive", "bash", false},
		{"empty string is not destructive", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w.isDestructiveTool(tt.tool); got != tt.want {
				t.Errorf("isDestructiveTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestGenerateGitDiff(t *testing.T) {
	// Create temp git repo
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial file
	os.WriteFile("test.txt", []byte("original\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Modify file
	os.WriteFile("test.txt", []byte("modified\n"), 0644)

	// Test diff generation
	diff, err := generateGitDiff("test.txt")
	if err != nil {
		t.Fatalf("generateGitDiff() error = %v", err)
	}

	if diff.FilePath != "test.txt" {
		t.Errorf("FilePath = %q, want %q", diff.FilePath, "test.txt")
	}

	if diff.DiffText == "" {
		t.Error("DiffText is empty")
	}

	if !contains(diff.DiffText, "-original") {
		t.Error("DiffText missing deletion line")
	}

	if !contains(diff.DiffText, "+modified") {
		t.Error("DiffText missing addition line")
	}

	if diff.Timestamp.IsZero() {
		t.Error("Timestamp not set")
	}
}

func TestGenerateGitDiff_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)
	os.WriteFile("test.txt", []byte("content"), 0644)

	_, err := generateGitDiff("test.txt")
	if err == nil {
		t.Error("Expected error for non-git repo, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && hasSubstr(s, substr))
}

func hasSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		name   string
		args   map[string]any
		want   string
		wantOk bool
	}{
		{
			name:   "nil args",
			args:   nil,
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty args",
			args:   map[string]any{},
			want:   "",
			wantOk: false,
		},
		{
			name:   "file_path key",
			args:   map[string]any{"file_path": "test.go"},
			want:   "test.go",
			wantOk: true,
		},
		{
			name:   "path key",
			args:   map[string]any{"path": "test.go"},
			want:   "test.go",
			wantOk: true,
		},
		{
			name:   "filepath key",
			args:   map[string]any{"filepath": "test.go"},
			want:   "test.go",
			wantOk: true,
		},
		{
			name:   "non-string value",
			args:   map[string]any{"file_path": 123},
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty string value",
			args:   map[string]any{"file_path": ""},
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := extractFilePath(tt.args)
			if got != tt.want || gotOk != tt.wantOk {
				t.Errorf("extractFilePath() = (%q, %v), want (%q, %v)",
					got, gotOk, tt.want, tt.wantOk)
			}
		})
	}
}
