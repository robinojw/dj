package appserver

type threadScoped struct {
	ThreadID string `json:"thread_id"`
}

// Turn represents a v2 turn object within notifications.
type Turn struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type turnNotification struct {
	threadScoped
	Turn Turn `json:"turn"`
}

// TurnStartedNotification is the params payload for turn/started.
type TurnStartedNotification struct {
	turnNotification
}

// TurnCompletedNotification is the params payload for turn/completed.
type TurnCompletedNotification struct {
	turnNotification
}
