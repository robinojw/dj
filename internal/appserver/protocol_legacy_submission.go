package appserver

import "encoding/json"

// Deprecated: Legacy outgoing operation format. Remove after client migration.

// ProtoSubmission is the legacy outgoing operation format.
type ProtoSubmission struct {
	ID string          `json:"id"`
	Op json.RawMessage `json:"op"`
}
