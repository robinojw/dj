package agents

import (
	"os"
	"os/exec"
	"testing"

	"github.com/robinojw/dj/internal/tools"
)

func TestIsMutatingTool_WithRegistry(t *testing.T) {
	registry := tools.NewDefaultRegistry(t.TempDir())
	w := &Worker{registry: registry}

	tests := []struct {
		name string
		tool string
		want bool
	}{
		{"edit_file mutates files", "edit_file", true},
		{"write_file mutates files", "write_file", true},
		{"str_replace mutates files", "str_replace", true},
		{"delete_file mutates files", "delete_file", true},
		{"read_file does not mutate", "read_file", false},
		{"list_dir does not mutate", "list_dir", false},
		{"run_tests does not mutate", "run_tests", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w.isMutatingTool(tt.tool); got != tt.want {
				t.Errorf("isMutatingTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestIsMutatingTool_UnregisteredReturnsFalse(t *testing.T) {
	registry := tools.NewDefaultRegistry(t.TempDir())
	w := &Worker{registry: registry}

	if w.isMutatingTool("some_mcp_tool") {
		t.Error("isMutatingTool for unregistered tool = true, want false")
	}
}

func TestIsMutatingTool_NilRegistry(t *testing.T) {
	w := &Worker{}
	if w.isMutatingTool("write_file") {
		t.Error("isMutatingTool with nil registry = true, want false")
	}
}

func TestIsMutatingTool_AnnotationsOnly(t *testing.T) {
	registry := tools.NewRegistry()
	registry.RegisterAnnotationsOnly("mcp_file_write", tools.ToolAnnotations{
		MutatesFiles:  true,
		FilePathParam: "path",
	})
	w := &Worker{registry: registry}

	if !w.isMutatingTool("mcp_file_write") {
		t.Error("isMutatingTool(mcp_file_write) = false, want true")
	}
}

func TestExtractToolFilePath_WithRegistry(t *testing.T) {
	registry := tools.NewDefaultRegistry(t.TempDir())
	w := &Worker{registry: registry}

	got, ok := w.extractToolFilePath("edit_file", map[string]any{"file_path": "main.go"})
	if !ok || got != "main.go" {
		t.Errorf("extractToolFilePath(edit_file) = (%q, %v), want (\"main.go\", true)", got, ok)
	}

	// Wrong key should fail even though value exists
	got, ok = w.extractToolFilePath("edit_file", map[string]any{"path": "main.go"})
	if ok {
		t.Errorf("extractToolFilePath(edit_file, path key) = (%q, true), want (\"\", false)", got)
	}
}

func TestExtractToolFilePath_AnnotationsOnly(t *testing.T) {
	registry := tools.NewRegistry()
	registry.RegisterAnnotationsOnly("mcp_write", tools.ToolAnnotations{
		MutatesFiles:  true,
		FilePathParam: "target",
	})
	w := &Worker{registry: registry}

	got, ok := w.extractToolFilePath("mcp_write", map[string]any{"target": "out.txt"})
	if !ok || got != "out.txt" {
		t.Errorf("extractToolFilePath(mcp_write) = (%q, %v), want (\"out.txt\", true)", got, ok)
	}
}

func TestExtractToolFilePath_FallbackForUnregistered(t *testing.T) {
	registry := tools.NewDefaultRegistry(t.TempDir())
	w := &Worker{registry: registry}

	got, ok := w.extractToolFilePath("unknown_tool", map[string]any{"path": "foo.go"})
	if !ok || got != "foo.go" {
		t.Errorf("extractToolFilePath(unknown) = (%q, %v), want (\"foo.go\", true)", got, ok)
	}
}

func TestGenerateGitDiff(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("test.txt", []byte("original\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	os.WriteFile("test.txt", []byte("modified\n"), 0644)

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
		{"nil args", nil, "", false},
		{"empty args", map[string]any{}, "", false},
		{"file_path key", map[string]any{"file_path": "test.go"}, "test.go", true},
		{"path key", map[string]any{"path": "test.go"}, "test.go", true},
		{"filepath key", map[string]any{"filepath": "test.go"}, "test.go", true},
		{"non-string value", map[string]any{"file_path": 123}, "", false},
		{"empty string value", map[string]any{"file_path": ""}, "", false},
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
