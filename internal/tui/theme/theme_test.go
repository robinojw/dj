package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme.Name != "tokyonight" {
		t.Errorf("expected theme name 'tokyonight', got %q", theme.Name)
	}

	// Verify all color fields are set
	colors := theme.Colors
	if colors.Background == "" {
		t.Error("background color should not be empty")
	}
	if colors.Foreground == "" {
		t.Error("foreground color should not be empty")
	}
	if colors.Primary == "" {
		t.Error("primary color should not be empty")
	}
	if colors.Error == "" {
		t.Error("error color should not be empty")
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary theme file
	tmpDir := t.TempDir()
	themePath := filepath.Join(tmpDir, "test-theme.json")

	testTheme := ThemeFile{
		Name: "test",
		Colors: ThemeColors{
			Background:  "#000000",
			Foreground:  "#ffffff",
			Primary:     "#ff0000",
			Secondary:   "#00ff00",
			Accent:      "#0000ff",
			Error:       "#ff0000",
			Warning:     "#ffff00",
			Success:     "#00ff00",
			Muted:       "#888888",
			Border:      "#444444",
			PanelBg:     "#111111",
			StatusBg:    "#222222",
			SelectionBg: "#333333",
			BadgeBg:     "#444444",
			BadgeFg:     "#555555",
		},
	}

	data, err := json.Marshal(testTheme)
	if err != nil {
		t.Fatalf("failed to marshal theme: %v", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		t.Fatalf("failed to write theme file: %v", err)
	}

	// Load the theme
	theme, err := LoadFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if theme.Name != "test" {
		t.Errorf("expected theme name 'test', got %q", theme.Name)
	}

	if theme.Colors.Background != "#000000" {
		t.Errorf("expected background #000000, got %s", theme.Colors.Background)
	}
}

func TestLoadFromFile_InvalidPath(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/theme.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	themePath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(themePath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	_, err := LoadFromFile(themePath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadBuiltin(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a built-in theme file
	testTheme := ThemeFile{
		Name: "custom",
		Colors: ThemeColors{
			Background: "#123456",
			Foreground: "#abcdef",
			Primary:    "#ff00ff",
			Secondary:  "#00ffff",
			Accent:     "#ffff00",
			Error:      "#ff0000",
			Warning:    "#ffa500",
			Success:    "#00ff00",
			Muted:      "#808080",
			Border:     "#404040",
			PanelBg:    "#101010",
			StatusBg:   "#202020",
			SelectionBg: "#303030",
			BadgeBg:    "#404040",
			BadgeFg:    "#505050",
		},
	}

	data, err := json.Marshal(testTheme)
	if err != nil {
		t.Fatalf("failed to marshal theme: %v", err)
	}

	themePath := filepath.Join(tmpDir, "custom.json")
	if err := os.WriteFile(themePath, data, 0644); err != nil {
		t.Fatalf("failed to write theme file: %v", err)
	}

	theme, err := LoadBuiltin("custom", tmpDir)
	if err != nil {
		t.Fatalf("LoadBuiltin failed: %v", err)
	}

	if theme.Name != "custom" {
		t.Errorf("expected theme name 'custom', got %q", theme.Name)
	}

	if theme.Colors.Background != "#123456" {
		t.Errorf("expected background #123456, got %s", theme.Colors.Background)
	}
}

func TestThemeStyles(t *testing.T) {
	theme := DefaultTheme()

	// Test that all style methods return non-nil styles
	tests := []struct {
		name  string
		style func() any
	}{
		{"PanelStyle", func() any { return theme.PanelStyle() }},
		{"StatusStyle", func() any { return theme.StatusStyle() }},
		{"BadgeStyle", func() any { return theme.BadgeStyle() }},
		{"AccentStyle", func() any { return theme.AccentStyle() }},
		{"ErrorStyle", func() any { return theme.ErrorStyle() }},
		{"SuccessStyle", func() any { return theme.SuccessStyle() }},
		{"MutedStyle", func() any { return theme.MutedStyle() }},
		{"SelectedStyle", func() any { return theme.SelectedStyle() }},
		{"PrimaryStyle", func() any { return theme.PrimaryStyle() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.style()
			if style == nil {
				t.Errorf("%s returned nil", tt.name)
			}
		})
	}
}
