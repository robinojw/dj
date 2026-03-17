package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) handleSessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return app, tea.Quit
	case tea.KeyEsc:
		app.closeSession()
		return app, nil
	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
		app.scrollSession(msg)
		return app, nil
	}
	return app, nil
}

func (app AppModel) openSession() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	thread, exists := app.store.Get(threadID)
	if !exists {
		return app, nil
	}

	session := NewSessionModel(thread)
	session.SetSize(app.width, app.height)
	app.session = &session
	app.focus = FocusSession
	return app, nil
}

func (app *AppModel) closeSession() {
	app.session = nil
	app.focus = FocusCanvas
}

func (app *AppModel) scrollSession(msg tea.KeyMsg) {
	if app.session == nil {
		return
	}

	switch msg.Type {
	case tea.KeyUp:
		app.session.viewport.ScrollUp(1)
	case tea.KeyDown:
		app.session.viewport.ScrollDown(1)
	case tea.KeyPgUp:
		app.session.viewport.HalfPageUp()
	case tea.KeyPgDown:
		app.session.viewport.HalfPageDown()
	}
}

func (app *AppModel) refreshSession() {
	if app.session == nil {
		return
	}
	app.session.Refresh()
}

func (app AppModel) handleThreadMessage(msg ThreadMessageMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendMessage(state.ChatMessage{
		ID:      msg.MessageID,
		Role:    msg.Role,
		Content: msg.Content,
	})
	app.refreshSession()
	return app, nil
}

func (app AppModel) handleThreadDelta(msg ThreadDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta(msg.MessageID, msg.Delta)
	app.refreshSession()
	return app, nil
}

func (app AppModel) handleCommandOutput(msg CommandOutputMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendOutput(msg.ExecID, msg.Data)
	app.refreshSession()
	return app, nil
}
