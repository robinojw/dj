package appserver

import "testing"

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
		{"AgentReasoning", EventAgentReasoning, "agent_reasoning"},
		{"AgentReasonDelta", EventAgentReasonDelta, "agent_reasoning_delta"},
		{"TokenCount", EventTokenCount, "token_count"},
		{"ExecApproval", EventExecApproval, "exec_command_request"},
		{"PatchApproval", EventPatchApproval, "patch_apply_request"},
		{"ReasonBreak", EventAgentReasonBreak, "agent_reasoning_section_break"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestOpConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"UserInput", OpUserInput, "user_input"},
		{"Interrupt", OpInterrupt, "interrupt"},
		{"ExecApproval", OpExecApproval, "exec_approval"},
		{"PatchApproval", OpPatchApproval, "patch_approval"},
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
