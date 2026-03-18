package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

func decodeV2ExecApproval(message appserver.JSONRPCMessage) tea.Msg {
	var request appserver.CommandApprovalRequest
	if err := json.Unmarshal(message.Params, &request); err != nil {
		return nil
	}
	return V2ExecApprovalMsg{
		RequestID: message.ID,
		ThreadID:  request.ThreadID,
		Command:   request.Command.Command,
		Cwd:       request.Command.Cwd,
	}
}

func decodeV2FileApproval(message appserver.JSONRPCMessage) tea.Msg {
	var request appserver.FileChangeApprovalRequest
	if err := json.Unmarshal(message.Params, &request); err != nil {
		return nil
	}
	return V2FileApprovalMsg{
		RequestID: message.ID,
		ThreadID:  request.ThreadID,
		Patch:     request.Patch,
	}
}

func decodeCollabSpawnEnd(raw json.RawMessage) tea.Msg {
	var event appserver.CollabSpawnEndEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil
	}
	return CollabSpawnMsg{
		SenderThreadID:   event.SenderThreadID,
		NewThreadID:      event.NewThreadID,
		NewAgentNickname: event.NewAgentNickname,
		NewAgentRole:     event.NewAgentRole,
		Status:           event.Status,
	}
}

func decodeCollabCloseEnd(raw json.RawMessage) tea.Msg {
	var event appserver.CollabCloseEndEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil
	}
	return CollabCloseMsg{
		SenderThreadID:   event.SenderThreadID,
		ReceiverThreadID: event.ReceiverThreadID,
		Status:           event.Status,
	}
}
