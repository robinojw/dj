package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight      = 1
	statusBarHeight   = 1
	viewSeparator     = "\n"
	colorDimGray      = "240"
	scrollIndicatorFg = "255"
)

func joinSections(sections ...string) string {
	return strings.Join(sections, viewSeparator)
}

func (app AppModel) renderBottomBar() string {
	if app.inputBarVisible {
		return app.inputBar.ViewWithWidth(app.width)
	}
	return app.statusBar.View()
}

func (app AppModel) View() string {
	title := app.header.View()
	status := app.renderBottomBar()

	if app.helpVisible {
		return joinSections(title, app.help.View(), status)
	}

	if app.menuVisible {
		return joinSections(title, app.menu.View(), status)
	}

	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0

	if hasPinned {
		return app.renderSplitView(title, status)
	}

	canvasHeight := app.height - headerHeight - statusBarHeight
	if canvasHeight < 1 {
		canvasHeight = 1
	}
	app.canvas.SetDimensions(app.width, canvasHeight)
	canvas := app.renderCanvas()
	return joinSections(title, canvas, status)
}

func (app AppModel) renderSplitView(title string, status string) string {
	canvasHeight := int(float64(app.height)*app.sessionPanel.SplitRatio()) - headerHeight - statusBarHeight
	if canvasHeight < 1 {
		canvasHeight = 1
	}
	app.canvas.SetDimensions(app.width, canvasHeight)
	canvas := app.renderCanvas()
	divider := app.renderDivider()
	panel := app.renderSessionPanel()
	return joinSections(title, canvas, divider, panel, status)
}

func (app AppModel) renderCanvas() string {
	app.canvas.SetPinnedIDs(app.sessionPanel.PinnedSessions())
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
	isEmptyAndExited := !hasVisibleContent && !ptySession.Running()
	if isEmptyAndExited {
		return fmt.Sprintf("[process exited: %d]", ptySession.ExitCode())
	}

	if ptySession.IsScrolledUp() {
		content = overlayScrollIndicator(content, ptySession.ScrollOffset())
	}

	return content
}

func overlayScrollIndicator(content string, linesBelow int) string {
	indicator := renderScrollIndicator(linesBelow)
	lines := strings.Split(content, viewSeparator)
	if len(lines) > 0 {
		lines[len(lines)-1] = indicator
	}
	return strings.Join(lines, viewSeparator)
}

func renderScrollIndicator(linesBelow int) string {
	text := fmt.Sprintf(" ↓ %d lines below ", linesBelow)
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(colorDimGray)).
		Foreground(lipgloss.Color(scrollIndicatorFg))
	return style.Render(text)
}

func (app AppModel) sessionPaneStyle(width int, height int, active bool) lipgloss.Style {
	borderColor := lipgloss.Color(colorDimGray)
	if active {
		borderColor = lipgloss.Color("39")
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - sessionPaneBorderSize).
		Height(height - sessionPaneBorderSize)
}
