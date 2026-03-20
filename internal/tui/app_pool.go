package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) handleSpawnRequest(msg SpawnRequestMsg) (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	agentID, err := app.pool.Spawn(msg.Persona, msg.Task, msg.SourceAgentID)
	if err != nil {
		app.statusBar.SetError(err.Error())
		return app, nil
	}

	app.store.Add(agentID, msg.Task)
	thread, threadExists := app.store.Get(agentID)
	if threadExists {
		thread.AgentProcessID = agentID
		thread.AgentRole = msg.Persona
		thread.ParentID = msg.SourceAgentID
	}

	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleAgentMessage(msg AgentMessageMsg) (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	targetAgent, exists := app.pool.Get(msg.TargetAgentID)
	if !exists {
		return app, nil
	}

	if targetAgent.Client == nil {
		return app, nil
	}

	targetAgent.Client.SendUserInput(msg.Content)
	return app, nil
}

func (app AppModel) handleAgentComplete(msg AgentCompleteMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.AgentID, state.StatusCompleted, "")
	app.store.UpdateActivity(msg.AgentID, "")
	return app, nil
}
