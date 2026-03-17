package appserver

import "testing"

func TestMethodConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ThreadStart", MethodThreadStart, "thread/start"},
		{"ThreadList", MethodThreadList, "thread/list"},
		{"ThreadArchive", MethodThreadArchive, "thread/archive"},
		{"TurnStart", MethodTurnStart, "turn/start"},
		{"TurnInterrupt", MethodTurnInterrupt, "turn/interrupt"},
		{"NotifyThreadStatus", NotifyThreadStatusChanged, "thread/status/changed"},
		{"NotifyItemStarted", NotifyItemStarted, "item/started"},
		{"NotifyItemCompleted", NotifyItemCompleted, "item/completed"},
		{"NotifyItemMessageDelta", NotifyItemMessageDelta, "item/agentMessage/delta"},
		{"NotifyTurnStarted", NotifyTurnStarted, "turn/started"},
		{"NotifyTurnCompleted", NotifyTurnCompleted, "turn/completed"},
		{"NotifyCommandOutput", NotifyCommandOutput, "command/output"},
		{"NotifyCommandFinished", NotifyCommandFinished, "command/finished"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}
