package agents

import (
	"os"
	"os/exec"
	"testing"
)

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
