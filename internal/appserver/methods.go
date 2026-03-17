package appserver

const (
	MethodThreadStart   = "thread/start"
	MethodThreadList    = "thread/list"
	MethodThreadArchive = "thread/archive"
	MethodTurnStart     = "turn/start"
	MethodTurnInterrupt = "turn/interrupt"
)

const (
	NotifyThreadStatusChanged = "thread/status/changed"
	NotifyItemStarted         = "item/started"
	NotifyItemCompleted       = "item/completed"
	NotifyItemMessageDelta    = "item/agentMessage/delta"
	NotifyTurnStarted         = "turn/started"
	NotifyTurnCompleted       = "turn/completed"
	NotifyCommandOutput       = "command/output"
	NotifyCommandFinished     = "command/finished"
)
