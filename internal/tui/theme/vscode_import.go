package theme

import (
	"encoding/json"
	"os"
)

// VSCodeTheme is a minimal representation of a VS Code color theme JSON.
type VSCodeTheme struct {
	Name   string            `json:"name"`
	Colors map[string]string `json:"colors"`
}

// ImportVSCodeTheme converts a VS Code theme JSON file into a Theme.
func ImportVSCodeTheme(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var vsc VSCodeTheme
	if err := json.Unmarshal(data, &vsc); err != nil {
		return nil, err
	}

	colors := ThemeColors{
		Background:  vscColor(vsc.Colors, "editor.background", "#1a1b26"),
		Foreground:  vscColor(vsc.Colors, "editor.foreground", "#c0caf5"),
		Primary:     vscColor(vsc.Colors, "focusBorder", "#7aa2f7"),
		Secondary:   vscColor(vsc.Colors, "button.background", "#bb9af7"),
		Accent:      vscColor(vsc.Colors, "textLink.foreground", "#7dcfff"),
		Error:       vscColor(vsc.Colors, "errorForeground", "#f7768e"),
		Warning:     vscColor(vsc.Colors, "editorWarning.foreground", "#e0af68"),
		Success:     vscColor(vsc.Colors, "terminal.ansiGreen", "#9ece6a"),
		Muted:       vscColor(vsc.Colors, "disabledForeground", "#565f89"),
		Border:      vscColor(vsc.Colors, "panel.border", "#3b4261"),
		PanelBg:     vscColor(vsc.Colors, "sideBar.background", "#1f2335"),
		StatusBg:    vscColor(vsc.Colors, "statusBar.background", "#16161e"),
		SelectionBg: vscColor(vsc.Colors, "editor.selectionBackground", "#283457"),
		BadgeBg:     vscColor(vsc.Colors, "badge.background", "#7aa2f7"),
		BadgeFg:     vscColor(vsc.Colors, "badge.foreground", "#1a1b26"),
	}

	return &Theme{
		Name:   vsc.Name,
		Colors: colors,
	}, nil
}

func vscColor(colors map[string]string, key, fallback string) string {
	if v, ok := colors[key]; ok && v != "" {
		return v
	}
	return fallback
}
