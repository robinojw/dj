package appserver

// ThreadStatusChangedNotification is the params payload for thread/status/changed.
type ThreadStatusChangedNotification struct {
	threadScoped
	Status string `json:"status"`
}

// AgentMessageDeltaNotification is the params payload for item/agentMessage/delta.
type AgentMessageDeltaNotification struct {
	threadScoped
	Delta string `json:"delta"`
}

// Item represents a v2 item object within notifications.
type Item struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ItemCompletedNotification is the params payload for item/completed.
type ItemCompletedNotification struct {
	threadScoped
	Item Item `json:"item"`
}
