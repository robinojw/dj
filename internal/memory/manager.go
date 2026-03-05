package memory

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager handles project-level (AGENTS.md) and user-level (memory.md) persistence.
type Manager struct {
	projectPath string // ./AGENTS.md
	userPath    string // ~/.config/codex-harness/memory.md
}

func NewManager(projectPath, userPath string) *Manager {
	return &Manager{
		projectPath: projectPath,
		userPath:    userPath,
	}
}

// DefaultManager creates a manager with standard paths.
func DefaultManager() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		projectPath: "AGENTS.md",
		userPath:    filepath.Join(home, ".config", "codex-harness", "memory.md"),
	}
}

// LoadContext reads both memory files and returns formatted context for injection.
func (m *Manager) LoadContext() string {
	project := readFileOrEmpty(m.projectPath)
	user := readFileOrEmpty(m.userPath)

	return fmt.Sprintf("<project_memory>\n%s\n</project_memory>\n\n<user_memory>\n%s\n</user_memory>",
		project, user)
}

// AppendUserMemory appends a line to the user memory file, creating it if needed.
func (m *Manager) AppendUserMemory(content string) error {
	dir := filepath.Dir(m.userPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	f, err := os.OpenFile(m.userPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s", content)
	return err
}

// ProjectPath returns the project memory file path.
func (m *Manager) ProjectPath() string {
	return m.projectPath
}

// UserPath returns the user memory file path.
func (m *Manager) UserPath() string {
	return m.userPath
}

func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
