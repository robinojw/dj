package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/orchestrator"
)

func (app AppModel) listenForPoolEvents() tea.Cmd {
	if app.pool == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-app.pool.Events()
		if !ok {
			return AppServerErrorMsg{Err: fmt.Errorf("pool events closed")}
		}
		return PoolEventMsg{
			AgentID: event.AgentID,
			Message: event.Message,
		}
	}
}

func (app AppModel) handlePoolEvent(msg PoolEventMsg) (tea.Model, tea.Cmd) {
	agent, exists := app.pool.Get(msg.AgentID)
	if !exists {
		return app, app.listenForPoolEvents()
	}

	tuiMsg := V2MessageToMsg(msg.Message)
	if tuiMsg == nil {
		return app, app.listenForPoolEvents()
	}

	deltaMsg, isDelta := tuiMsg.(V2AgentDeltaMsg)
	if !isDelta {
		updated, innerCmd := app.Update(tuiMsg)
		resultApp := updated.(AppModel)
		return resultApp, tea.Batch(innerCmd, resultApp.listenForPoolEvents())
	}

	agent.Parser.Feed(deltaMsg.Delta)
	commands := agent.Parser.Flush()
	return app.processCommands(msg.AgentID, commands, tuiMsg)
}

func (app AppModel) processCommands(agentID string, commands []orchestrator.DJCommand, originalMsg tea.Msg) (tea.Model, tea.Cmd) {
	updated, innerCmd := app.Update(originalMsg)
	resultApp := updated.(AppModel)
	cmds := []tea.Cmd{innerCmd}

	for _, command := range commands {
		resultApp, cmds = applyCommand(resultApp, agentID, command, cmds)
	}

	cmds = append(cmds, resultApp.listenForPoolEvents())
	return resultApp, tea.Batch(cmds...)
}

func applyCommand(app AppModel, agentID string, command orchestrator.DJCommand, cmds []tea.Cmd) (AppModel, []tea.Cmd) {
	switch command.Action {
	case "spawn":
		return applySpawnCommand(app, agentID, command, cmds)
	case "message":
		return applyMessageCommand(app, agentID, command, cmds)
	case "complete":
		return applyCompleteCommand(app, agentID, command, cmds)
	}
	return app, cmds
}

func applySpawnCommand(app AppModel, agentID string, command orchestrator.DJCommand, cmds []tea.Cmd) (AppModel, []tea.Cmd) {
	spawnMsg := SpawnRequestMsg{
		SourceAgentID: agentID,
		Persona:       command.Persona,
		Task:          command.Task,
	}
	spawnUpdated, spawnCmd := app.handleSpawnRequest(spawnMsg)
	return spawnUpdated.(AppModel), append(cmds, spawnCmd)
}

func applyMessageCommand(app AppModel, agentID string, command orchestrator.DJCommand, cmds []tea.Cmd) (AppModel, []tea.Cmd) {
	agentMsg := AgentMessageMsg{
		SourceAgentID: agentID,
		TargetAgentID: command.Target,
		Content:       command.Content,
	}
	msgUpdated, msgCmd := app.handleAgentMessage(agentMsg)
	return msgUpdated.(AppModel), append(cmds, msgCmd)
}

func applyCompleteCommand(app AppModel, agentID string, command orchestrator.DJCommand, cmds []tea.Cmd) (AppModel, []tea.Cmd) {
	completeMsg := AgentCompleteMsg{
		AgentID: agentID,
		Content: command.Content,
	}
	completeUpdated, completeCmd := app.handleAgentComplete(completeMsg)
	return completeUpdated.(AppModel), append(cmds, completeCmd)
}
