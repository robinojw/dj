package appserver

import "testing"

func TestMethodConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ThreadCreate", MethodThreadCreate, "thread/create"},
		{"ThreadList", MethodThreadList, "thread/list"},
		{"ThreadDelete", MethodThreadDelete, "thread/delete"},
		{"ThreadSendMessage", MethodThreadSendMessage, "thread/sendMessage"},
		{"CommandExec", MethodCommandExec, "command/exec"},
		{"NotifyThreadStatus", NotifyThreadStatusChanged, "thread/status/changed"},
		{"NotifyThreadMessage", NotifyThreadMessageCreated, "thread/message/created"},
		{"NotifyMessageDelta", NotifyThreadMessageDelta, "thread/message/delta"},
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
