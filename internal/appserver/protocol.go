package appserver

import "encoding/json"

// ProtoEvent is an incoming event from the codex proto stream.
// Format: {"id":"<correlation-id>","msg":{"type":"<event-type>",...}}
type ProtoEvent struct {
	ID  string          `json:"id"`
	Msg json.RawMessage `json:"msg"`
}

// EventHeader extracts just the type field from a ProtoEvent.Msg payload.
type EventHeader struct {
	Type string `json:"type"`
}

// ProtoSubmission is an outgoing operation sent to codex proto.
// Format: {"id":"<correlation-id>","op":{"type":"<op-type>",...}}
type ProtoSubmission struct {
	ID string          `json:"id"`
	Op json.RawMessage `json:"op"`
}
