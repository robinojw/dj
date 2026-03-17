package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

type MessageSender interface {
	Send(msg tea.Msg)
}

func WireEventBridge(router *appserver.NotificationRouter, sender MessageSender) {
	router.OnThreadStatusChanged(func(params appserver.ThreadStatusChanged) {
		sender.Send(ThreadStatusMsg{
			ThreadID: params.ThreadID,
			Status:   params.Status,
			Title:    params.Title,
		})
	})

	router.OnItemStarted(func(params appserver.ItemStarted) {
		sender.Send(ThreadMessageMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.ItemID,
			Role:      params.Role,
			Content:   "",
		})
	})

	router.OnItemMessageDelta(func(params appserver.ItemMessageDelta) {
		sender.Send(ThreadDeltaMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.ItemID,
			Delta:     params.Delta,
		})
	})

	router.OnCommandOutput(func(params appserver.CommandOutput) {
		sender.Send(CommandOutputMsg{
			ThreadID: params.ThreadID,
			ExecID:   params.ExecID,
			Data:     params.Data,
		})
	})

	router.OnCommandFinished(func(params appserver.CommandFinished) {
		sender.Send(CommandFinishedMsg{
			ThreadID: params.ThreadID,
			ExecID:   params.ExecID,
			ExitCode: params.ExitCode,
		})
	})
}
