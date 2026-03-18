package appserver

// CollabWaitingBeginEvent is the params payload for collab/agentWaiting/begin.
type CollabWaitingBeginEvent struct {
	collabBase
	ReceiverThreadIDs []string `json:"receiver_thread_ids"`
}

// CollabWaitingEndEvent is the params payload for collab/agentWaiting/end.
type CollabWaitingEndEvent struct {
	collabBase
	Statuses map[string]string `json:"statuses"`
}
