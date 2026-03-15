package tui

import (
	"os"
	"os/exec"
	"strings"
)

func resolveGitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "no-branch"
	}
	return strings.TrimSpace(string(out))
}

func resolveWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}
