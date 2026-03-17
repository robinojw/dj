package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

type MessageSender interface {
	Send(msg tea.Msg)
}

func WireEventBridge(router *appserver.EventRouter, sender MessageSender) {
	router.OnSessionConfigured(func(event appserver.SessionConfigured) {
		sender.Send(AppServerConnectedMsg{
			SessionID: event.SessionID,
			Model:     event.Model,
		})
	})

	router.OnAgentMessageDelta(func(event appserver.AgentMessageDelta) {
		sender.Send(ThreadDeltaMsg{
			Delta: event.Delta,
		})
	})

	router.OnExecCommandBegin(func(event appserver.ExecCommandBegin) {
		sender.Send(CommandOutputMsg{
			ExecID: event.ExecID,
			Data:   "$ " + event.Command + "\n",
		})
	})

	router.OnExecCommandOutputDelta(func(event appserver.ExecCommandOutputDelta) {
		sender.Send(CommandOutputMsg{
			ExecID: event.ExecID,
			Data:   event.Delta,
		})
	})

	router.OnExecCommandEnd(func(event appserver.ExecCommandEnd) {
		sender.Send(CommandFinishedMsg{
			ExecID:   event.ExecID,
			ExitCode: event.ExitCode,
		})
	})

	router.OnError(func(event appserver.ServerError) {
		sender.Send(AppServerErrorMsg{
			Err: &appserver.RPCError{Message: event.Message},
		})
	})
}
