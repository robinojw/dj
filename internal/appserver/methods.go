package appserver

const (
	MethodThreadCreate      = "thread/create"
	MethodThreadList        = "thread/list"
	MethodThreadDelete      = "thread/delete"
	MethodThreadSendMessage = "thread/sendMessage"
	MethodCommandExec       = "command/exec"
)

const (
	NotifyThreadStatusChanged  = "thread/status/changed"
	NotifyThreadMessageCreated = "thread/message/created"
	NotifyThreadMessageDelta   = "thread/message/delta"
	NotifyCommandOutput        = "command/output"
	NotifyCommandFinished      = "command/finished"
)
