package appserver

import "testing"

func TestOpConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"UserTurn", OpUserTurn, "user_turn"},
		{"Interrupt", OpInterrupt, "interrupt"},
		{"ExecApproval", OpExecApproval, "exec_approval"},
		{"Shutdown", OpShutdown, "shutdown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"SessionConfigured", EventSessionConfigured, "session_configured"},
		{"TaskStarted", EventTaskStarted, "task_started"},
		{"TaskComplete", EventTaskComplete, "task_complete"},
		{"AgentMessage", EventAgentMessage, "agent_message"},
		{"AgentMessageDelta", EventAgentMessageDelta, "agent_message_delta"},
		{"ExecCommandBegin", EventExecCommandBegin, "exec_command_begin"},
		{"ExecCommandOutputDelta", EventExecCommandOutputDelta, "exec_command_output_delta"},
		{"ExecCommandEnd", EventExecCommandEnd, "exec_command_end"},
		{"ExecApprovalRequest", EventExecApprovalRequest, "exec_approval_request"},
		{"Error", EventError, "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}
