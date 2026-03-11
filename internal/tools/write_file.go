package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteFileHandler returns a ToolHandler for write_file.
// Creates parent directories if needed and backs up existing files before overwriting.
// Backups are stored in a .dj-backups directory within the workspace root.
func WriteFileHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		filePath, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("write_file: missing required argument 'file_path'")
		}

		content, ok := stringArg(args, "content")
		if !ok {
			return "", fmt.Errorf("write_file: missing required argument 'content'")
		}

		absPath, err := safePath(workspaceRoot, filePath)
		if err != nil {
			return "", fmt.Errorf("write_file: %w", err)
		}

		// Create parent directories
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("write_file: create parent dirs: %w", err)
		}

		// Backup existing file to .dj-backups/
		var origMode os.FileMode = 0644
		if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
			origMode = info.Mode()
			backupDir := filepath.Join(workspaceRoot, ".dj-backups")
			if err := os.MkdirAll(backupDir, 0755); err == nil {
				backupName := fmt.Sprintf("%s.%d", filepath.Base(absPath), time.Now().UnixMilli())
				backupPath := filepath.Join(backupDir, backupName)
				if data, err := os.ReadFile(absPath); err == nil {
					_ = os.WriteFile(backupPath, data, origMode)
				}
			}
		}

		if err := os.WriteFile(absPath, []byte(content), origMode); err != nil {
			return "", fmt.Errorf("write_file: %w", err)
		}

		return fmt.Sprintf("Wrote %d bytes to %s", len(content), filePath), nil
	}
}
