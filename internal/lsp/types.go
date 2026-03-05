package lsp

// Diagnostic represents an LSP diagnostic (error, warning, etc.).
type Diagnostic struct {
	File     string
	Line     int
	Column   int
	Severity string // "error", "warning", "info", "hint"
	Message  string
	Source   string // "gopls", "typescript-language-server", etc.
}

// ServerConfig holds the configuration to launch an LSP server.
type ServerConfig struct {
	Language string // "go", "typescript", "python"
	Command  string // "gopls", "typescript-language-server --stdio"
	Args     []string
}

// DetectedServer is the result of auto-detection.
type DetectedServer struct {
	Config   ServerConfig
	RootPath string
}
