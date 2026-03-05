package theme

import "github.com/charmbracelet/lipgloss"

// AdaptiveColor returns a color that adapts to light/dark terminal backgrounds.
func AdaptiveColor(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

// Subtle returns a muted version of the theme's foreground.
func (t *Theme) Subtle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Muted))
}

// Highlight returns a style using the primary color on the panel background.
func (t *Theme) Highlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Primary)).
		Background(lipgloss.Color(t.Colors.PanelBg)).
		Bold(true)
}

// BorderedBox returns a padded box with rounded border.
func (t *Theme) BorderedBox(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Colors.Border)).
		Padding(1, 2)
}

// ProgressBarColors returns foreground colors for a progress bar
// that shifts from green to yellow to red.
func (t *Theme) ProgressBarColors(pct float64) string {
	switch {
	case pct < 60:
		return t.Colors.Success
	case pct < 85:
		return t.Colors.Warning
	default:
		return t.Colors.Error
	}
}
