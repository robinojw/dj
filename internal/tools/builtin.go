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

type builtinSchema struct {
	name        string
	description string
	parameters  string
}

var builtinSchemas = []builtinSchema{
	{
		name:        "read_file",
		description: "Read a file's contents with line numbers. Returns numbered lines from the file within the workspace.",
		parameters: `{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file (relative to workspace root or absolute)"},
				"offset":    {"type": "integer", "description": "Line offset to start reading from (0-based). Defaults to 0."},
				"limit":     {"type": "integer", "description": "Maximum number of lines to read. Defaults to all lines."}
			},
			"required": ["file_path"],
			"additionalProperties": false
		}`,
	},
	{
		name:        "write_file",
		description: "Write content to a file, creating parent directories if needed. Backs up existing files before overwriting.",
		parameters: `{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file (relative to workspace root or absolute)"},
				"content":   {"type": "string", "description": "The full content to write to the file"}
			},
			"required": ["file_path", "content"],
			"additionalProperties": false
		}`,
	},
	{
		name:        "edit_file",
		description: "Edit a file by replacing an exact string match. Supports whitespace-tolerant matching if exact match fails.",
		parameters: `{
			"type": "object",
			"properties": {
				"file_path":  {"type": "string", "description": "Path to the file to edit"},
				"old_string": {"type": "string", "description": "The exact text to find and replace"},
				"new_string": {"type": "string", "description": "The replacement text. Omit or empty to delete the matched text."}
			},
			"required": ["file_path", "old_string"],
			"additionalProperties": false
		}`,
	},
	{
		name:        "str_replace",
		description: "Replace a string in a file (alias for edit_file). Finds old_string and replaces with new_string.",
		parameters: `{
			"type": "object",
			"properties": {
				"file_path":  {"type": "string", "description": "Path to the file to edit"},
				"old_string": {"type": "string", "description": "The exact text to find and replace"},
				"new_string": {"type": "string", "description": "The replacement text. Omit or empty to delete the matched text."}
			},
			"required": ["file_path", "old_string"],
			"additionalProperties": false
		}`,
	},
	{
		name:        "delete_file",
		description: "Delete a file from the workspace. Cannot delete directories.",
		parameters: `{
			"type": "object",
			"properties": {
				"file_path": {"type": "string", "description": "Path to the file to delete"}
			},
			"required": ["file_path"],
			"additionalProperties": false
		}`,
	},
	{
		name:        "list_dir",
		description: "List files and directories at a given path. Shows file sizes and directory indicators.",
		parameters: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Directory path to list (relative to workspace root). Defaults to workspace root."}
			},
			"additionalProperties": false
		}`,
	},
	{
		name:        "run_tests",
		description: "Run Go tests and return structured results with pass/fail status and output.",
		parameters: `{
			"type": "object",
			"properties": {
				"package": {"type": "string", "description": "Go package pattern to test. Defaults to './...' (all packages)."},
				"run":     {"type": "string", "description": "Regex filter for test names (passed as -run flag)."}
			},
			"additionalProperties": false
		}`,
	},
}

func registerBuiltinSchemas(r *ToolRegistry) {
	for _, s := range builtinSchemas {
		r.RegisterSchema(s.name, ToolSchema{
			Description: s.description,
			Parameters:  json.RawMessage(s.parameters),
		})
	}
}
