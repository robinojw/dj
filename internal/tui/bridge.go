package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

// ProtoEventToMsg converts a codex proto event into a Bubble Tea message.
func ProtoEventToMsg(event appserver.ProtoEvent) tea.Msg {
	var header appserver.EventHeader
	if err := json.Unmarshal(event.Msg, &header); err != nil {
		return nil
	}

	switch header.Type {
	case appserver.EventSessionConfigured:
		return decodeSessionConfigured(event.Msg)
	case appserver.EventTaskStarted:
		return TaskStartedMsg{}
	case appserver.EventTaskComplete:
		return decodeTaskComplete(event.Msg)
	case appserver.EventAgentMessageDelta:
		return decodeAgentDelta(event.Msg)
	case appserver.EventAgentMessage:
		return decodeAgentMessage(event.Msg)
	case appserver.EventAgentReasonDelta:
		return decodeReasoningDelta(event.Msg)
	case appserver.EventExecApproval:
		return decodeExecApproval(event)
	case appserver.EventPatchApproval:
		return decodePatchApproval(event)
	}
	return nil
}

func decodeSessionConfigured(raw json.RawMessage) tea.Msg {
	var config appserver.SessionConfigured
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil
	}
	return SessionConfiguredMsg{
		SessionID: config.SessionID,
		Model:     config.Model,
	}
}

func decodeTaskComplete(raw json.RawMessage) tea.Msg {
	var complete appserver.TaskComplete
	if err := json.Unmarshal(raw, &complete); err != nil {
		return nil
	}
	return TaskCompleteMsg{LastMessage: complete.LastAgentMessage}
}

func decodeAgentDelta(raw json.RawMessage) tea.Msg {
	var delta appserver.AgentDelta
	if err := json.Unmarshal(raw, &delta); err != nil {
		return nil
	}
	return AgentDeltaMsg{Delta: delta.Delta}
}

func decodeAgentMessage(raw json.RawMessage) tea.Msg {
	var msg appserver.AgentMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil
	}
	return AgentMessageCompletedMsg{Message: msg.Message}
}

func decodeReasoningDelta(raw json.RawMessage) tea.Msg {
	var delta appserver.AgentDelta
	if err := json.Unmarshal(raw, &delta); err != nil {
		return nil
	}
	return AgentReasoningDeltaMsg{Delta: delta.Delta}
}

func decodeExecApproval(event appserver.ProtoEvent) tea.Msg {
	var req appserver.ExecCommandRequest
	if err := json.Unmarshal(event.Msg, &req); err != nil {
		return nil
	}
	return ExecApprovalRequestMsg{
		EventID: event.ID,
		Command: req.Command,
		Cwd:     req.Cwd,
	}
}

func decodePatchApproval(event appserver.ProtoEvent) tea.Msg {
	var req appserver.PatchApplyRequest
	if err := json.Unmarshal(event.Msg, &req); err != nil {
		return nil
	}
	return PatchApprovalRequestMsg{
		EventID: event.ID,
		Patch:   req.Patch,
	}
}
