package tui

import "github.com/robinojw/dj/internal/appserver"

type MessageSender interface {
	Send(msg any)
}

func WireEventBridge(router *appserver.NotificationRouter, sender MessageSender) {
	router.OnThreadStatusChanged(func(params appserver.ThreadStatusChanged) {
		sender.Send(ThreadStatusMsg{
			ThreadID: params.ThreadID,
			Status:   params.Status,
			Title:    params.Title,
		})
	})

	router.OnThreadMessageCreated(func(params appserver.ThreadMessageCreated) {
		sender.Send(ThreadMessageMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.MessageID,
			Role:      params.Role,
			Content:   params.Content,
		})
	})

	router.OnThreadMessageDelta(func(params appserver.ThreadMessageDelta) {
		sender.Send(ThreadDeltaMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.MessageID,
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
