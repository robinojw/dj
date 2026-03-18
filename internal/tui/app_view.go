package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

func (app AppModel) View() string {
	title := titleStyle.Render("DJ — Codex TUI Visualizer")
	status := app.statusBar.View()

	if app.helpVisible {
		return title + "\n" + app.help.View() + "\n" + status
	}

	if app.menuVisible {
		return title + "\n" + app.menu.View() + "\n" + status
	}

	canvas := app.renderCanvas()
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0

	if !hasPinned {
		return title + "\n" + canvas + "\n" + status
	}

	divider := app.renderDivider()
	panel := app.renderSessionPanel()
	return title + "\n" + canvas + "\n" + divider + "\n" + panel + "\n" + status
}

func (app AppModel) renderCanvas() string {
	canvas := app.canvas.View()
	if app.canvasMode == CanvasModeTree {
		treeView := app.tree.View()
		return lipgloss.JoinHorizontal(lipgloss.Top, treeView+"  ", canvas)
	}
	return canvas
}

func (app AppModel) renderDivider() string {
	pinned := app.sessionPanel.PinnedSessions()
	activeIdx := app.sessionPanel.ActivePaneIdx()

	labels := make([]string, len(pinned))
	for index, threadID := range pinned {
		thread, exists := app.store.Get(threadID)
		if exists {
			labels[index] = thread.Title
		} else {
			labels[index] = threadID
		}
	}
	return renderDividerBar(labels, activeIdx, app.width)
}

func (app AppModel) renderSessionPanel() string {
	pinned := app.sessionPanel.PinnedSessions()
	if len(pinned) == 0 {
		return ""
	}

	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight

	if app.sessionPanel.Zoomed() {
		return app.renderZoomedSession(panelHeight)
	}
	return app.renderSideBySideSessions(pinned, panelHeight)
}

func (app AppModel) renderZoomedSession(panelHeight int) string {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return ""
	}

	content := app.renderPTYContent(activeID)
	style := app.sessionPaneStyle(app.width, panelHeight, true)
	return style.Render(content)
}

func (app AppModel) renderSideBySideSessions(pinned []string, panelHeight int) string {
	count := len(pinned)
	sessionWidth := app.width / count

	panes := make([]string, count)
	for index, threadID := range pinned {
		content := app.renderPTYContent(threadID)
		isActive := index == app.sessionPanel.ActivePaneIdx() && app.focusPane == FocusPaneSession
		style := app.sessionPaneStyle(sessionWidth, panelHeight, isActive)
		panes[index] = style.Render(content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panes...)
}

func (app AppModel) renderPTYContent(threadID string) string {
	ptySession, exists := app.ptySessions[threadID]
	if !exists {
		return ""
	}

	content := ptySession.Render()
	hasVisibleContent := strings.TrimSpace(content) != ""
	if hasVisibleContent {
		return content
	}

	if !ptySession.Running() {
		return fmt.Sprintf("[process exited: %d]", ptySession.ExitCode())
	}
	return content
}

func (app AppModel) sessionPaneStyle(width int, height int, active bool) lipgloss.Style {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("39")
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - sessionPaneBorderSize).
		Height(height - sessionPaneBorderSize)
}
