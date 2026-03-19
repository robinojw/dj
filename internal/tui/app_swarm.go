package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

const inputBarPromptSuffix = ": "

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
	app.inputBar = NewInputBarModel("Task for " + persona.Name + inputBarPromptSuffix)
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
	agentID := extractAgentID(item.Label)
	app.pendingTargetAgentID = agentID
	app.inputBar = NewInputBarModel("Message to " + item.Label + inputBarPromptSuffix)
	app.inputBarVisible = true
	app.inputBarIntent = IntentSendMessage
	return app, nil
}

func (app AppModel) sendMessageToAgent() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	agents := app.pool.All()
	if len(agents) == 0 {
		app.statusBar.SetError("No active agents")
		return app, nil
	}

	items := buildAgentMenuItems(agents)
	app.menu = NewMenuModel("Message Agent", items)
	app.menuVisible = true
	app.menuIntent = MenuIntentAgentPicker
	return app, nil
}

func buildAgentMenuItems(agents []*pool.AgentProcess) []MenuItem {
	items := make([]MenuItem, 0, len(agents))
	for _, agent := range agents {
		label := agent.ID
		if agent.Persona != nil {
			label = agent.Persona.Name + " (" + agent.ID + ")"
		}
		items = append(items, MenuItem{
			Label: label,
			Key:   rune(agent.ID[0]),
		})
	}
	return items
}

func extractAgentID(label string) string {
	parenStart := strings.LastIndex(label, "(")
	parenEnd := strings.LastIndex(label, ")")
	hasParen := parenStart != -1 && parenEnd > parenStart
	if hasParen {
		return label[parenStart+1 : parenEnd]
	}
	return label
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
