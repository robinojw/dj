package theme

import (
	tui "github.com/grindlemire/go-tui"
)

// mustHexColor parses a hex color string, returning DefaultColor on error.
func mustHexColor(hex string) tui.Color {
	c, err := tui.HexColor(hex)
	if err != nil {
		return tui.DefaultColor()
	}
	return c
}

// Style helpers that convert ThemeColors to go-tui styles.

func (t *Theme) TuiPrimaryStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Primary))
}

func (t *Theme) TuiAccentStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Accent))
}

func (t *Theme) TuiErrorStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Error))
}

func (t *Theme) TuiSuccessStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Success))
}

func (t *Theme) TuiWarningStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Warning))
}

func (t *Theme) TuiMutedStyle() tui.Style {
	return tui.NewStyle().Foreground(mustHexColor(t.Colors.Muted))
}

// Color helpers for use in go-tui element options.

func (t *Theme) TuiBorderColor() tui.Color {
	return mustHexColor(t.Colors.Border)
}

func (t *Theme) TuiAccentColor() tui.Color {
	return mustHexColor(t.Colors.Accent)
}

func (t *Theme) TuiPrimaryColor() tui.Color {
	return mustHexColor(t.Colors.Primary)
}

func (t *Theme) TuiMutedColor() tui.Color {
	return mustHexColor(t.Colors.Muted)
}

func (t *Theme) TuiErrorColor() tui.Color {
	return mustHexColor(t.Colors.Error)
}

func (t *Theme) TuiSuccessColor() tui.Color {
	return mustHexColor(t.Colors.Success)
}
