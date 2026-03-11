package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxReadBytes = 2 * 1024 * 1024 // 2 MB

// ReadFileHandler returns a ToolHandler for read_file with path traversal prevention.
// workspaceRoot is the directory that all paths must resolve within.
func ReadFileHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		filePath, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("read_file: missing required argument 'file_path'")
		}

		absPath, err := safePath(workspaceRoot, filePath)
		if err != nil {
			return "", fmt.Errorf("read_file: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("read_file: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("read_file: path is a directory, use list_dir instead")
		}
		if info.Size() > maxReadBytes {
			return "", fmt.Errorf("read_file: file too large (%d bytes, max %d)", info.Size(), maxReadBytes)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("read_file: %w", err)
		}

		lines := strings.Split(string(data), "\n")

		offset := intArg(args, "offset", 0)
		limit := intArg(args, "limit", len(lines))

		if offset < 0 {
			offset = 0
		}
		if offset > len(lines) {
			offset = len(lines)
		}
		end := offset + limit
		if end > len(lines) {
			end = len(lines)
		}

		selected := lines[offset:end]

		// Format with line numbers
		var sb strings.Builder
		for i, line := range selected {
			fmt.Fprintf(&sb, "%d\t%s\n", offset+i+1, line)
		}
		return sb.String(), nil
	}
}

// safePath resolves filePath relative to root and prevents path traversal.
// Resolves symlinks to prevent symlink-based escapes from the workspace.
func safePath(root, filePath string) (string, error) {
	rootCleaned := filepath.Clean(root)
	// Resolve symlinks in the root so all comparisons use canonical paths.
	if resolved, err := filepath.EvalSymlinks(rootCleaned); err == nil {
		rootCleaned = resolved
	}

	var abs string
	if filepath.IsAbs(filePath) {
		abs = filepath.Clean(filePath)
	} else {
		abs = filepath.Clean(filepath.Join(rootCleaned, filePath))
	}

	// Resolve symlinks in the target path if it exists.
	// For new files, resolve the parent directory to catch symlinked parents.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	} else if resolved, err := filepath.EvalSymlinks(filepath.Dir(abs)); err == nil {
		abs = filepath.Join(resolved, filepath.Base(abs))
	}

	if !strings.HasPrefix(abs, rootCleaned+string(filepath.Separator)) && abs != rootCleaned {
		return "", fmt.Errorf("path %q is outside workspace root", filePath)
	}

	return abs, nil
}

// stringArg extracts a string argument from the args map.
func stringArg(args map[string]any, key string) (string, bool) {
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// intArg extracts an integer argument from the args map, returning defaultVal if not found.
func intArg(args map[string]any, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	return defaultVal
}
