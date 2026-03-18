package appserver

import "encoding/json"

// Deprecated: Legacy incoming event format. Remove after client migration.

// ProtoEvent is the legacy incoming event format.
type ProtoEvent struct {
	ID  string          `json:"id"`
	Msg json.RawMessage `json:"msg"`
}

// EventHeader extracts just the type field from a ProtoEvent.Msg payload.
type EventHeader struct {
	Type string `json:"type"`
}
