package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ThemeColors holds the color palette for a theme.
type ThemeColors struct {
	Background  string `json:"background"`
	Foreground  string `json:"foreground"`
	Primary     string `json:"primary"`
	Secondary   string `json:"secondary"`
	Accent      string `json:"accent"`
	Error       string `json:"error"`
	Warning     string `json:"warning"`
	Success     string `json:"success"`
	Muted       string `json:"muted"`
	Border      string `json:"border"`
	PanelBg     string `json:"panel_bg"`
	StatusBg    string `json:"status_bg"`
	SelectionBg string `json:"selection_bg"`
	BadgeBg     string `json:"badge_bg"`
	BadgeFg     string `json:"badge_fg"`
}

// ThemeFile is the JSON structure of a theme file.
type ThemeFile struct {
	Name   string      `json:"name"`
	Colors ThemeColors `json:"colors"`
}

// Theme provides color palettes for the TUI.
type Theme struct {
	Name   string
	Colors ThemeColors
}

// LoadFromFile reads a theme from a JSON file.
func LoadFromFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tf ThemeFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, err
	}
	return &Theme{Name: tf.Name, Colors: tf.Colors}, nil
}

// LoadBuiltin loads a theme by name from the built-in themes directory.
func LoadBuiltin(name string, themesDir string) (*Theme, error) {
	path := filepath.Join(themesDir, name+".json")
	return LoadFromFile(path)
}

// DefaultTheme returns the built-in tokyonight theme as a fallback.
func DefaultTheme() *Theme {
	return &Theme{
		Name: "tokyonight",
		Colors: ThemeColors{
			Background:  "#1a1b26",
			Foreground:  "#c0caf5",
			Primary:     "#7aa2f7",
			Secondary:   "#bb9af7",
			Accent:      "#7dcfff",
			Error:       "#f7768e",
			Warning:     "#e0af68",
			Success:     "#9ece6a",
			Muted:       "#565f89",
			Border:      "#3b4261",
			PanelBg:     "#1f2335",
			StatusBg:    "#16161e",
			SelectionBg: "#283457",
			BadgeBg:     "#7aa2f7",
			BadgeFg:     "#1a1b26",
		},
	}
}
