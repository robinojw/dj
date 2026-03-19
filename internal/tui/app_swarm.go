package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) showPersonaPicker() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	personas := app.pool.Personas()
	if len(personas) == 0 {
		app.statusBar.SetError("No personas available")
		return app, nil
	}

	items := buildPersonaMenuItems(personas)
	app.menu = NewMenuModel("Spawn Persona Agent", items)
	app.menuVisible = true
	app.menuIntent = MenuIntentPersonaPicker
	return app, nil
}

func buildPersonaMenuItems(personas map[string]roster.PersonaDefinition) []MenuItem {
	items := make([]MenuItem, 0, len(personas))
	for _, persona := range personas {
		items = append(items, MenuItem{
			Label: persona.Name,
			Key:   rune(persona.ID[0]),
		})
	}
	return items
}

func (app AppModel) dispatchPersonaPick(item MenuItem) (tea.Model, tea.Cmd) {
	persona := app.findPersonaByName(item.Label)
	if persona == nil {
		return app, nil
	}

	app.pendingPersonaID = persona.ID
	app.inputBar = NewInputBarModel("Task for " + persona.Name + ": ")
	app.inputBarVisible = true
	app.inputBarIntent = IntentSpawnTask
	return app, nil
}

func (app AppModel) findPersonaByName(name string) *roster.PersonaDefinition {
	if app.pool == nil {
		return nil
	}
	for _, persona := range app.pool.Personas() {
		if persona.Name == name {
			return &persona
		}
	}
	return nil
}

func (app AppModel) dispatchAgentPick(item MenuItem) (tea.Model, tea.Cmd) {
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
