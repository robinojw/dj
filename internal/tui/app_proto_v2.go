package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) handleThreadStarted(msg ThreadStartedMsg) (tea.Model, tea.Cmd) {
	isSubAgent := msg.SourceType == appserver.SourceTypeSubAgent
	if isSubAgent {
		app.store.AddSubAgent(msg.ThreadID, msg.AgentNickname, msg.ParentID, msg.AgentNickname, msg.AgentRole, msg.Depth)
	} else {
		app.store.Add(msg.ThreadID, msg.ThreadID)
	}
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleTurnStarted(msg TurnStartedMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.ThreadID, state.StatusActive, "")
	return app, nil
}

func (app AppModel) handleTurnCompleted(msg TurnCompletedMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.ThreadID, state.StatusCompleted, "")
	return app, nil
}

func (app AppModel) handleV2AgentDelta(msg V2AgentDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta("", msg.Delta)
	return app, nil
}

func (app AppModel) handleCollabSpawn(msg CollabSpawnMsg) (tea.Model, tea.Cmd) {
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleCollabClose(msg CollabCloseMsg) (tea.Model, tea.Cmd) {
	agentStatus := mapAgentStatusToDJ(msg.Status)
	app.store.UpdateStatus(msg.ReceiverThreadID, agentStatus, "")
	return app, nil
}

func (app AppModel) handleThreadStatusChanged(msg ThreadStatusChangedMsg) (tea.Model, tea.Cmd) {
	agentStatus := mapAgentStatusToDJ(msg.Status)
	app.store.UpdateStatus(msg.ThreadID, agentStatus, "")
	return app, nil
}

func (app AppModel) handleV2ExecApproval(msg V2ExecApprovalMsg) (tea.Model, tea.Cmd) {
	if app.client != nil {
		app.client.SendApproval(msg.RequestID, true)
	}
	return app, nil
}

func (app AppModel) handleV2FileApproval(msg V2FileApprovalMsg) (tea.Model, tea.Cmd) {
	if app.client != nil {
		app.client.SendApproval(msg.RequestID, true)
	}
	return app, nil
}

func mapAgentStatusToDJ(agentStatus string) string {
	statusMap := map[string]string{
		appserver.AgentStatusPendingInit: state.StatusIdle,
		appserver.AgentStatusRunning:     state.StatusActive,
		appserver.AgentStatusInterrupted: state.StatusIdle,
		appserver.AgentStatusCompleted:   state.StatusCompleted,
		appserver.AgentStatusErrored:     state.StatusError,
		appserver.AgentStatusShutdown:    state.StatusCompleted,
	}
	djStatus, exists := statusMap[agentStatus]
	if !exists {
		return state.StatusIdle
	}
	return djStatus
}
