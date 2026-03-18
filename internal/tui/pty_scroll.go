package tui

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

func renderScrolledViewport(
	scrollbackLines []uv.Line,
	screenLines []string,
	viewportHeight int,
	scrollOffset int,
) []string {
	allLines := make([]string, 0, len(scrollbackLines)+len(screenLines))

	for _, line := range scrollbackLines {
		allLines = append(allLines, line.Render())
	}
	allLines = append(allLines, screenLines...)

	totalLines := len(allLines)
	end := totalLines - scrollOffset
	if end < 0 {
		end = 0
	}
	start := end - viewportHeight
	if start < 0 {
		start = 0
	}
	if end > totalLines {
		end = totalLines
	}

	visible := allLines[start:end]

	for len(visible) < viewportHeight {
		visible = append([]string{""}, visible...)
	}

	return visible
}

func renderScrolledOutput(
	scrollbackLines []uv.Line,
	screenLines []string,
	viewportHeight int,
	scrollOffset int,
) string {
	lines := renderScrolledViewport(scrollbackLines, screenLines, viewportHeight, scrollOffset)
	return strings.Join(lines, "\n")
}
