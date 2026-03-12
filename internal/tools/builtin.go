package tools

import "encoding/json"

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

	registerBuiltinSchemas(r)

	return r
}

func registerBuiltinSchemas(r *ToolRegistry) {
	r.RegisterSchema("read_file", ToolSchema{
		Description: "Read a file's contents with line numbers. Returns numbered lines from the file within the workspace.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file (relative to workspace root or absolute)"},
				"offset":    {"type": "integer", "description": "Line offset to start reading from (0-based). Defaults to 0."},
				"limit":     {"type": "integer", "description": "Maximum number of lines to read. Defaults to all lines."}
			},
			"required": ["file_path"],
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("write_file", ToolSchema{
		Description: "Write content to a file, creating parent directories if needed. Backs up existing files before overwriting.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file (relative to workspace root or absolute)"},
				"content":   {"type": "string", "description": "The full content to write to the file"}
			},
			"required": ["file_path", "content"],
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("edit_file", ToolSchema{
		Description: "Edit a file by replacing an exact string match. Supports whitespace-tolerant matching if exact match fails.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"file_path":  {"type": "string", "description": "Path to the file to edit"},
				"old_string": {"type": "string", "description": "The exact text to find and replace"},
				"new_string": {"type": "string", "description": "The replacement text. Omit or empty to delete the matched text."}
			},
			"required": ["file_path", "old_string"],
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("str_replace", ToolSchema{
		Description: "Replace a string in a file (alias for edit_file). Finds old_string and replaces with new_string.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"file_path":  {"type": "string", "description": "Path to the file to edit"},
				"old_string": {"type": "string", "description": "The exact text to find and replace"},
				"new_string": {"type": "string", "description": "The replacement text. Omit or empty to delete the matched text."}
			},
			"required": ["file_path", "old_string"],
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("delete_file", ToolSchema{
		Description: "Delete a file from the workspace. Cannot delete directories.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file to delete"}
			},
			"required": ["file_path"],
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("list_dir", ToolSchema{
		Description: "List files and directories at a given path. Shows file sizes and directory indicators.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Directory path to list (relative to workspace root). Defaults to workspace root."}
			},
			"additionalProperties": false
		}`),
	})

	r.RegisterSchema("run_tests", ToolSchema{
		Description: "Run Go tests and return structured results with pass/fail status and output.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"package": {"type": "string", "description": "Go package pattern to test. Defaults to './...' (all packages)."},
				"run":     {"type": "string", "description": "Regex filter for test names (passed as -run flag)."}
			},
			"additionalProperties": false
		}`),
	})
}
