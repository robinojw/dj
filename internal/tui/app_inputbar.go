package tui

import tea "github.com/charmbracelet/bubbletea"

func (app AppModel) handleInputBarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return app.dismissInputBar()
	case tea.KeyEnter:
		return app.submitInputBar()
	case tea.KeyBackspace:
		app.inputBar.DeleteRune()
		return app, nil
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			app.inputBar.InsertRune(r)
		}
		return app, nil
	}
	return app, nil
}

func (app AppModel) dismissInputBar() (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	app.pendingPersonaID = ""
	app.pendingTargetAgentID = ""
	app.inputBar.Reset()
	return app, nil
}

func (app AppModel) submitInputBar() (tea.Model, tea.Cmd) {
	value := app.inputBar.Value()
	isEmpty := value == ""
	if isEmpty {
		return app.dismissInputBar()
	}

	switch app.inputBarIntent {
	case IntentSpawnTask:
		return app.executeSpawn(value)
	case IntentSendMessage:
		return app.executeSendMessage(value)
	}
	return app.dismissInputBar()
}

func (app AppModel) executeSpawn(task string) (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	personaID := app.pendingPersonaID
	app.pendingPersonaID = ""
	app.inputBar.Reset()

	if app.pool == nil {
		return app, nil
	}

	agentID, spawnErr := app.pool.Spawn(personaID, task, "")
	if spawnErr != nil {
		app.statusBar.SetError(spawnErr.Error())
		return app, nil
	}

	app.store.Add(agentID, task)
	app.store.UpdateStatus(agentID, "active", "")
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) executeSendMessage(content string) (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	targetID := app.pendingTargetAgentID
	app.pendingTargetAgentID = ""
	app.inputBar.Reset()

	if app.pool == nil {
		return app, nil
	}

	targetAgent, exists := app.pool.Get(targetID)
	if !exists {
		app.statusBar.SetError("Agent not found")
		return app, nil
	}

	if targetAgent.Client != nil {
		targetAgent.Client.SendUserInput(content)
	}
	return app, nil
}
