package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewDefaultRegistry creates a ToolRegistry pre-loaded with all built-in tool handlers.
func NewDefaultRegistry(workspaceRoot string) *ToolRegistry {
	r := NewRegistry()

	r.Register("read_file", ReadFileHandler(workspaceRoot), ToolAnnotations{
		ReadOnly:   true,
		Idempotent: true,
	})

	r.Register("write_file", WriteFileHandler(workspaceRoot), ToolAnnotations{
		Destructive:   true,
		MutatesFiles:  true,
		FilePathParam: "file_path",
	})

	r.Register("edit_file", EditFileHandler(workspaceRoot), ToolAnnotations{
		Destructive:   true,
		MutatesFiles:  true,
		FilePathParam: "file_path",
	})

	r.Register("str_replace", EditFileHandler(workspaceRoot), ToolAnnotations{
		Destructive:   true,
		MutatesFiles:  true,
		FilePathParam: "file_path",
	})

	r.Register("delete_file", DeleteFileHandler(workspaceRoot), ToolAnnotations{
		Destructive:   true,
		MutatesFiles:  true,
		FilePathParam: "file_path",
	})

	r.Register("list_dir", ListDirHandler(workspaceRoot), ToolAnnotations{
		ReadOnly:   true,
		Idempotent: true,
	})

	r.Register("run_tests", RunTestsHandler(workspaceRoot), ToolAnnotations{
		ReadOnly: true,
	})

	return r
}

// ReadFileHandler returns a handler that reads a file's contents.
func ReadFileHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		path, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("missing file_path argument")
		}
		data, err := os.ReadFile(resolvePath(root, path))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

// WriteFileHandler returns a handler that writes content to a file.
func WriteFileHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		path, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("missing file_path argument")
		}
		content, ok := stringArg(args, "content")
		if !ok {
			return "", fmt.Errorf("missing content argument")
		}
		resolved := resolvePath(root, path)
		if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(resolved, []byte(content), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("Wrote %s", path), nil
	}
}

// EditFileHandler returns a handler that performs string replacement in a file.
func EditFileHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		path, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("missing file_path argument")
		}
		oldStr, ok := stringArg(args, "old_string")
		if !ok {
			return "", fmt.Errorf("missing old_string argument")
		}
		newStr, _ := stringArg(args, "new_string")
		resolved := resolvePath(root, path)
		data, err := os.ReadFile(resolved)
		if err != nil {
			return "", err
		}
		replaced := strings.Replace(string(data), oldStr, newStr, 1)
		if err := os.WriteFile(resolved, []byte(replaced), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("Edited %s", path), nil
	}
}

// DeleteFileHandler returns a handler that deletes a file.
func DeleteFileHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		path, ok := stringArg(args, "file_path")
		if !ok {
			return "", fmt.Errorf("missing file_path argument")
		}
		if err := os.Remove(resolvePath(root, path)); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deleted %s", path), nil
	}
}

// ListDirHandler returns a handler that lists directory contents.
func ListDirHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		path, ok := stringArg(args, "path")
		if !ok {
			path = "."
		}
		entries, err := os.ReadDir(resolvePath(root, path))
		if err != nil {
			return "", err
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return strings.Join(names, "\n"), nil
	}
}

// RunTestsHandler returns a handler stub for running tests.
func RunTestsHandler(root string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		return "run_tests: not implemented via local handler", nil
	}
}

// stringArg extracts a string argument from a map.
func stringArg(args map[string]any, key string) (string, bool) {
	if args == nil {
		return "", false
	}
	val, ok := args[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok && str != ""
}

// resolvePath resolves a potentially relative path against the workspace root.
func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
