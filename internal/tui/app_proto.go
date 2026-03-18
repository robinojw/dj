package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

const (
	activityThinking      = "Thinking..."
	activityApplyingPatch = "Applying patch..."
	activityRunningPrefix = "Running: "
	activitySnippetMaxLen = 40
)

type protoEventMsg struct {
	Event appserver.ProtoEvent
}

func (app AppModel) connectClient() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := app.client.Start(ctx); err != nil {
			return AppServerErrorMsg{Err: err}
		}
		go app.client.ReadLoop(func(event appserver.ProtoEvent) {
			app.events <- event
		})
		return nil
	}
}

func (app AppModel) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-app.events
		if !ok {
			return AppServerErrorMsg{Err: fmt.Errorf("connection closed")}
		}
		return protoEventMsg{Event: event}
	}
}

func (app AppModel) handleProtoEvent(event appserver.ProtoEvent) (tea.Model, tea.Cmd) {
	tuiMsg := ProtoEventToMsg(event)
	if tuiMsg == nil {
		return app, app.listenForEvents()
	}
	updated, innerCmd := app.Update(tuiMsg)
	resultApp := updated.(AppModel)
	return resultApp, tea.Batch(innerCmd, resultApp.listenForEvents())
}

func (app AppModel) handleSessionConfigured(msg SessionConfiguredMsg) (tea.Model, tea.Cmd) {
	app.sessionID = msg.SessionID
	title := msg.Model
	if title == "" {
		title = "Codex Session"
	}
	app.store.Add(msg.SessionID, title)
	app.statusBar.SetConnected(true)
	app.statusBar.SetThreadCount(len(app.store.All()))
	return app, nil
}

func (app AppModel) handleTaskStarted() (tea.Model, tea.Cmd) {
	if app.sessionID == "" {
		return app, nil
	}
	app.store.UpdateStatus(app.sessionID, state.StatusActive, "")
	app.store.UpdateActivity(app.sessionID, activityThinking)
	messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())
	app.currentMessageID = messageID
	thread, exists := app.store.Get(app.sessionID)
	if exists {
		thread.AppendMessage(state.ChatMessage{
			ID:      messageID,
			Role:    "assistant",
			Content: "",
		})
	}
	return app, nil
}

func (app AppModel) handleAgentDelta(msg AgentDeltaMsg) (tea.Model, tea.Cmd) {
	missingContext := app.sessionID == "" || app.currentMessageID == ""
	if missingContext {
		return app, nil
	}
	thread, exists := app.store.Get(app.sessionID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta(app.currentMessageID, msg.Delta)
	snippet := latestMessageSnippet(thread, app.currentMessageID)
	app.store.UpdateActivity(app.sessionID, snippet)
	return app, nil
}

func (app AppModel) handleReasoningDelta() (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, activityThinking)
	}
	return app, nil
}

func (app AppModel) handleAgentMessageCompleted() (tea.Model, tea.Cmd) {
	app.currentMessageID = ""
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, "")
	}
	return app, nil
}

func (app AppModel) handleTaskComplete() (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateStatus(app.sessionID, state.StatusCompleted, "")
		app.store.UpdateActivity(app.sessionID, "")
	}
	app.currentMessageID = ""
	return app, nil
}

func (app AppModel) handleExecApproval(msg ExecApprovalRequestMsg) (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		activity := activityRunningPrefix + msg.Command
		app.store.UpdateActivity(app.sessionID, activity)
	}
	if app.client != nil {
		app.client.SendApproval(msg.EventID, appserver.OpExecApproval, true)
	}
	return app, nil
}

func (app AppModel) handlePatchApproval(msg PatchApprovalRequestMsg) (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, activityApplyingPatch)
	}
	if app.client != nil {
		app.client.SendApproval(msg.EventID, appserver.OpPatchApproval, true)
	}
	return app, nil
}

func latestMessageSnippet(thread *state.ThreadState, messageID string) string {
	for index := range thread.Messages {
		if thread.Messages[index].ID != messageID {
			continue
		}
		content := thread.Messages[index].Content
		if len(content) <= activitySnippetMaxLen {
			return content
		}
		return content[len(content)-activitySnippetMaxLen:]
	}
	return ""
}
