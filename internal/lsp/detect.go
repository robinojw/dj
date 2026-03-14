package lsp

import (
	"os"
	"os/exec"
)

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
		markerExists := fileExists(root + "/" + marker)
		if !markerExists {
			continue
		}
		if _, err := exec.LookPath(cfg.Command); err != nil {
			continue
		}
		return &DetectedServer{Config: cfg, RootPath: root}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
