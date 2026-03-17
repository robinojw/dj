package appserver

import "encoding/json"

// Submission is a client-to-server message in the Codex proto protocol.
type Submission struct {
	ID string          `json:"id"`
	Op json.RawMessage `json:"op"`
}

// Event is a server-to-client message in the Codex proto protocol.
type Event struct {
	ID  string          `json:"id"`
	Msg json.RawMessage `json:"msg"`
}

// EventHeader extracts just the type discriminator from an event message.
type EventHeader struct {
	Type string `json:"type"`
}

// RPCError is an error returned by the server.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}
