package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) showPersonaPicker() (tea.Model, tea.Cmd) {
	return app, nil
}

func (app AppModel) sendMessageToAgent() (tea.Model, tea.Cmd) {
	return app, nil
}

func (app AppModel) killAgent() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}
	threadID := app.canvas.SelectedThreadID()
	agent, exists := app.pool.GetByThreadID(threadID)
	if !exists {
		return app, nil
	}
	app.pool.StopAgent(agent.ID)
	app.store.UpdateStatus(threadID, state.StatusCompleted, "")
	return app, nil
}

func (app AppModel) toggleSwarmView() (tea.Model, tea.Cmd) {
	return app, nil
}
