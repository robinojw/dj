package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxListEntries = 1000

// ListDirHandler returns a ToolHandler for list_dir scoped to the workspace root.
func ListDirHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		dirPath, ok := stringArg(args, "path")
		if !ok {
			dirPath = "."
		}

		absPath, err := safePath(workspaceRoot, dirPath)
		if err != nil {
			return "", fmt.Errorf("list_dir: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("list_dir: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("list_dir: %s is not a directory", dirPath)
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			return "", fmt.Errorf("list_dir: %w", err)
		}

		var sb strings.Builder
		count := 0
		for _, entry := range entries {
			if count >= maxListEntries {
				fmt.Fprintf(&sb, "... truncated (%d entries total)\n", len(entries))
				break
			}

			name := entry.Name()
			// Make path relative to workspace root for display
			relPath, err := filepath.Rel(workspaceRoot, filepath.Join(absPath, name))
			if err != nil {
				relPath = name
			}

			if entry.IsDir() {
				fmt.Fprintf(&sb, "%s/\n", relPath)
			} else {
				info, err := entry.Info()
				if err != nil {
					fmt.Fprintf(&sb, "%s\n", relPath)
				} else {
					fmt.Fprintf(&sb, "%s (%d bytes)\n", relPath, info.Size())
				}
			}
			count++
		}

		if count == 0 {
			return "(empty directory)", nil
		}

		return sb.String(), nil
	}
}
