package appserver

type collabInteractionBase struct {
	collabBase
	ReceiverThreadID string `json:"receiver_thread_id"`
}

// CollabInteractionBeginEvent is the params payload for collab/agentInteraction/begin.
type CollabInteractionBeginEvent struct {
	collabInteractionBase
	Prompt string `json:"prompt,omitempty"`
}

// CollabInteractionEndEvent is the params payload for collab/agentInteraction/end.
type CollabInteractionEndEvent struct {
	collabInteractionBase
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
	Status                string `json:"status"`
}
