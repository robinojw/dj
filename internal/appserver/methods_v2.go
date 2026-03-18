package appserver

// V2 server notification methods.
const (
	MethodThreadStarted       = "thread/started"
	MethodThreadStatusChanged = "thread/status/changed"
	MethodThreadClosed        = "thread/closed"
	MethodTurnStarted         = "turn/started"
	MethodTurnCompleted       = "turn/completed"
	MethodItemStarted         = "item/started"
	MethodItemCompleted       = "item/completed"
	MethodAgentMessageDelta   = "item/agentMessage/delta"
	MethodTokenUsageUpdated   = "thread/tokenUsage/updated"
	MethodExecOutputDelta     = "item/commandExecution/outputDelta"
	MethodErrorNotification   = "error"
)

// V2 server request methods (require response).
const (
	MethodExecApproval = "item/commandExecution/requestApproval"
	MethodFileApproval = "item/fileChange/requestApproval"
)

// V2 client request methods (outgoing).
const (
	MethodInitialize    = "initialize"
	MethodThreadStart   = "thread/start"
	MethodTurnStart     = "turn/start"
	MethodTurnInterrupt = "turn/interrupt"
)

// V2 collaboration notification methods.
const (
	MethodCollabSpawnBegin       = "collab/agentSpawn/begin"
	MethodCollabSpawnEnd         = "collab/agentSpawn/end"
	MethodCollabInteractionBegin = "collab/agentInteraction/begin"
	MethodCollabInteractionEnd   = "collab/agentInteraction/end"
	MethodCollabWaitingBegin     = "collab/agentWaiting/begin"
	MethodCollabWaitingEnd       = "collab/agentWaiting/end"
	MethodCollabCloseBegin       = "collab/agentClose/begin"
	MethodCollabCloseEnd         = "collab/agentClose/end"
	MethodCollabResumeBegin      = "collab/agentResume/begin"
	MethodCollabResumeEnd        = "collab/agentResume/end"
)
