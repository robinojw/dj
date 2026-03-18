package appserver

const (
	EventSessionConfigured = "session_configured"
	EventTaskStarted       = "task_started"
	EventTaskComplete      = "task_complete"
	EventAgentMessage      = "agent_message"
	EventAgentMessageDelta = "agent_message_delta"
	EventAgentReasoning    = "agent_reasoning"
	EventAgentReasonDelta  = "agent_reasoning_delta"
	EventTokenCount        = "token_count"
	EventExecApproval      = "exec_command_request"
	EventPatchApproval     = "patch_apply_request"
	EventAgentReasonBreak  = "agent_reasoning_section_break"
)

const (
	OpUserInput     = "user_input"
	OpInterrupt     = "interrupt"
	OpExecApproval  = "exec_approval"
	OpPatchApproval = "patch_approval"
	OpShutdown      = "shutdown"
)
