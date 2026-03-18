package appserver

type collabReceiverBase struct {
	collabBase
	ReceiverThreadID      string `json:"receiver_thread_id"`
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
}

// CollabCloseBeginEvent is the params payload for collab/agentClose/begin.
type CollabCloseBeginEvent struct {
	collabReceiverBase
}

// CollabResumeBeginEvent is the params payload for collab/agentResume/begin.
type CollabResumeBeginEvent struct {
	collabReceiverBase
}
