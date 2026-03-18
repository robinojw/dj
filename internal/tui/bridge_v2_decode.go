package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

func decodeThreadStarted(raw json.RawMessage) tea.Msg {
	var notification appserver.ThreadStartedNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	thread := notification.Thread
	return ThreadStartedMsg{
		ThreadID:      thread.ID,
		Status:        thread.Status,
		SourceType:    thread.Source.Type,
		ParentID:      thread.Source.ParentThreadID,
		Depth:         thread.Source.Depth,
		AgentNickname: thread.Source.AgentNickname,
		AgentRole:     thread.Source.AgentRole,
	}
}

func decodeTurnStarted(raw json.RawMessage) tea.Msg {
	var notification appserver.TurnStartedNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	return TurnStartedMsg{
		ThreadID: notification.ThreadID,
		TurnID:   notification.Turn.ID,
	}
}

func decodeTurnCompleted(raw json.RawMessage) tea.Msg {
	var notification appserver.TurnCompletedNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	return TurnCompletedMsg{
		ThreadID: notification.ThreadID,
		TurnID:   notification.Turn.ID,
	}
}

func decodeV2AgentDelta(raw json.RawMessage) tea.Msg {
	var notification appserver.AgentMessageDeltaNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	return V2AgentDeltaMsg{
		ThreadID: notification.ThreadID,
		Delta:    notification.Delta,
	}
}

func decodeThreadStatusChanged(raw json.RawMessage) tea.Msg {
	var notification appserver.ThreadStatusChangedNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	return ThreadStatusChangedMsg{
		ThreadID: notification.ThreadID,
		Status:   notification.Status,
	}
}
