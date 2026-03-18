package appserver

type collabReceiverEndEvent struct {
	collabReceiverBase
	Status string `json:"status"`
}

// CollabCloseEndEvent is the params payload for collab/agentClose/end.
type CollabCloseEndEvent struct {
	collabReceiverEndEvent
}

// CollabResumeEndEvent is the params payload for collab/agentResume/end.
type CollabResumeEndEvent struct {
	collabReceiverEndEvent
}
