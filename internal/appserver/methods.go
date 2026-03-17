package appserver

const (
	OpUserTurn        = "user_turn"
	OpInterrupt       = "interrupt"
	OpExecApproval    = "exec_approval"
	OpPatchApproval   = "patch_approval"
	OpShutdown        = "shutdown"
	OpGetHistoryEntry = "get_history_entry_request"
)

const (
	EventSessionConfigured      = "session_configured"
	EventTaskStarted            = "task_started"
	EventTaskComplete           = "task_complete"
	EventTurnAborted            = "turn_aborted"
	EventAgentMessage           = "agent_message"
	EventAgentMessageDelta      = "agent_message_delta"
	EventExecCommandBegin       = "exec_command_begin"
	EventExecCommandOutputDelta = "exec_command_output_delta"
	EventExecCommandEnd         = "exec_command_end"
	EventExecApprovalRequest    = "exec_approval_request"
	EventPatchApplyBegin        = "patch_apply_begin"
	EventPatchApplyEnd          = "patch_apply_end"
	EventError                  = "error"
	EventWarning                = "warning"
	EventShutdownComplete       = "shutdown_complete"
)
