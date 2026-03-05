package lsp

import "os"

// knownServers maps language indicators to LSP server commands.
var knownServers = map[string]ServerConfig{
	"go.mod": {
		Language: "go",
		Command:  "gopls",
		Args:     []string{"serve"},
	},
	"tsconfig.json": {
		Language: "typescript",
		Command:  "typescript-language-server",
		Args:     []string{"--stdio"},
	},
	"package.json": {
		Language: "typescript",
		Command:  "typescript-language-server",
		Args:     []string{"--stdio"},
	},
	"pyproject.toml": {
		Language: "python",
		Command:  "pylsp",
		Args:     nil,
	},
	"setup.py": {
		Language: "python",
		Command:  "pylsp",
		Args:     nil,
	},
}

// Detect scans the project root for known language indicators
// and returns the first matching LSP server config.
func Detect(root string) *DetectedServer {
	for marker, cfg := range knownServers {
		if _, err := os.Stat(root + "/" + marker); err == nil {
			return &DetectedServer{Config: cfg, RootPath: root}
		}
	}
	return nil
}
