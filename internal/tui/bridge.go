package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

// ProtoEventToMsg converts a JSON-RPC message into a Bubble Tea message.
// Routes on legacy event types extracted from Params for backward compatibility.
func ProtoEventToMsg(message appserver.JsonRpcMessage) tea.Msg {
	if message.Params == nil {
		return nil
	}

	var header appserver.EventHeader
	if err := json.Unmarshal(message.Params, &header); err != nil {
		return nil
	}

	return routeLegacyEvent(header.Type, message)
}

func routeLegacyEvent(eventType string, message appserver.JsonRpcMessage) tea.Msg {
	switch eventType {
	case appserver.EventSessionConfigured:
		return decodeSessionConfigured(message.Params)
	case appserver.EventTaskStarted:
		return TaskStartedMsg{}
	case appserver.EventTaskComplete:
		return decodeTaskComplete(message.Params)
	case appserver.EventAgentMessageDelta:
		return decodeAgentDelta(message.Params)
	case appserver.EventAgentMessage:
		return decodeAgentMessage(message.Params)
	case appserver.EventAgentReasonDelta:
		return decodeReasoningDelta(message.Params)
	case appserver.EventExecApproval:
		return decodeExecApproval(message)
	case appserver.EventPatchApproval:
		return decodePatchApproval(message)
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
	var message appserver.AgentMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		return nil
	}
	return AgentMessageCompletedMsg{Message: message.Message}
}

func decodeReasoningDelta(raw json.RawMessage) tea.Msg {
	var delta appserver.AgentDelta
	if err := json.Unmarshal(raw, &delta); err != nil {
		return nil
	}
	return AgentReasoningDeltaMsg{Delta: delta.Delta}
}

func decodeExecApproval(message appserver.JsonRpcMessage) tea.Msg {
	var request appserver.ExecCommandRequest
	if err := json.Unmarshal(message.Params, &request); err != nil {
		return nil
	}
	return ExecApprovalRequestMsg{
		EventID: message.ID,
		Command: request.Command,
		Cwd:     request.Cwd,
	}
}

func decodePatchApproval(message appserver.JsonRpcMessage) tea.Msg {
	var request appserver.PatchApplyRequest
	if err := json.Unmarshal(message.Params, &request); err != nil {
		return nil
	}
	return PatchApprovalRequestMsg{
		EventID: message.ID,
		Patch:   request.Patch,
	}
}
