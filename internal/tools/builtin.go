package tools

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
