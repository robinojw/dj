package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.helpVisible {
		return app.handleHelpKey(msg)
	}

	if app.inputBarVisible {
		return app.handleInputBarKey(msg)
	}

	if app.menuVisible {
		return app.handleMenuKey(msg)
	}

	if result, model, cmd := app.handlePrefix(msg); result {
		return model, cmd
	}

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

	return app.handleCanvasKey(msg)
}

func (app AppModel) handlePrefix(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	prefixResult := app.prefix.HandleKey(msg)
	switch prefixResult {
	case PrefixWaiting:
		return true, app, nil
	case PrefixComplete:
		model, cmd := app.handlePrefixAction()
		return true, model, cmd
	case PrefixCancelled:
		return true, app, nil
	}
	return false, app, nil
}

func (app AppModel) handleCanvasKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyEnter:
		return app.openSession()
	case tea.KeyTab:
		return app.switchToSessionPanel()
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}

func (app AppModel) handleRune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		app.toggleCanvasMode()
	case "n":
		return app, app.createThread()
	case "?":
		app.helpVisible = !app.helpVisible
	case " ":
		return app.togglePin()
	case "k":
		return app.killSession()
	case "p":
		return app.showPersonaPicker()
	case "m":
		return app.sendMessageToAgent()
	case "s":
		return app.toggleSwarmView()
	case "K":
		return app.killAgent()
	}
	return app, nil
}

func (app AppModel) createThread() tea.Cmd {
	*app.sessionCounter++
	counter := *app.sessionCounter
	return func() tea.Msg {
		return ThreadCreatedMsg{
			ThreadID: fmt.Sprintf("session-%d", counter),
			Title:    fmt.Sprintf("Session %d", counter),
		}
	}
}

func (app AppModel) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isToggle := msg.Type == tea.KeyRunes && msg.String() == "?"
	isEsc := msg.Type == tea.KeyEsc
	shouldDismissHelp := isToggle || isEsc
	if shouldDismissHelp {
		app.helpVisible = false
	}
	return app, nil
}

func (app *AppModel) toggleCanvasMode() {
	if app.canvasMode == CanvasModeGrid {
		app.canvasMode = CanvasModeTree
		return
	}
	app.canvasMode = CanvasModeGrid
}

func (app AppModel) switchToSessionPanel() (tea.Model, tea.Cmd) {
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		return app, nil
	}
	app.focusPane = FocusPaneSession
	return app, nil
}

func (app AppModel) handleArrow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.canvasMode == CanvasModeTree {
		app.handleTreeArrow(msg)
		return app, nil
	}
	app.handleCanvasArrow(msg)
	return app, nil
}

func (app *AppModel) handleTreeArrow(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyDown:
		app.tree.MoveDown()
	case tea.KeyUp:
		app.tree.MoveUp()
	}
}

func (app *AppModel) handleCanvasArrow(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRight:
		app.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		app.canvas.MoveLeft()
	case tea.KeyDown:
		app.canvas.MoveDown()
	case tea.KeyUp:
		app.canvas.MoveUp()
	}
}
