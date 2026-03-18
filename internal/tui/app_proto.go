package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

const (
	activityThinking      = "Thinking..."
	activityApplyingPatch = "Applying patch..."
	activityRunningPrefix = "Running: "
	activitySnippetMaxLen = 40
)

type jsonRPCEventMsg struct {
	Message appserver.JSONRPCMessage
}

func (app AppModel) connectClient() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := app.client.Start(ctx); err != nil {
			return AppServerErrorMsg{Err: err}
		}
		go app.client.ReadLoop(func(message appserver.JSONRPCMessage) {
			app.events <- message
		})
		return nil
	}
}

func (app AppModel) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		message, ok := <-app.events
		if !ok {
			return AppServerErrorMsg{Err: fmt.Errorf("connection closed")}
		}
		return jsonRPCEventMsg{Message: message}
	}
}

func (app AppModel) handleProtoEvent(message appserver.JSONRPCMessage) (tea.Model, tea.Cmd) {
	tuiMsg := V2MessageToMsg(message)
	if tuiMsg == nil {
		return app, app.listenForEvents()
	}
	updated, innerCmd := app.Update(tuiMsg)
	resultApp := updated.(AppModel)
	return resultApp, tea.Batch(innerCmd, resultApp.listenForEvents())
}
