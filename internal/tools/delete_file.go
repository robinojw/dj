package tools

import (
	"context"
	"fmt"
	"os"
)

// DeleteFileHandler returns a ToolHandler for delete_file with path traversal prevention.
func DeleteFileHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		filePath, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("delete_file: missing required argument 'file_path'")
		}

		absPath, err := safePath(workspaceRoot, filePath)
		if err != nil {
			return "", fmt.Errorf("delete_file: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("delete_file: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("delete_file: cannot delete directory %s", filePath)
		}

		if err := os.Remove(absPath); err != nil {
			return "", fmt.Errorf("delete_file: %w", err)
		}

		return fmt.Sprintf("Deleted %s", filePath), nil
	}
}
