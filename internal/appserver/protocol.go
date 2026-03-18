package appserver

import "encoding/json"

// JsonRpcMessage represents a JSON-RPC 2.0 message (notification, request, or response).
type JsonRpcMessage struct {
	JsonRpc string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// IsRequest returns true if this message is a server-to-client request.
func (message JsonRpcMessage) IsRequest() bool {
	return message.ID != "" && message.Method != ""
}

// IsResponse returns true if this message is a response to a client request.
func (message JsonRpcMessage) IsResponse() bool {
	return message.ID != "" && message.Method == ""
}

// IsNotification returns true if this message is a server notification.
func (message JsonRpcMessage) IsNotification() bool {
	return message.ID == "" && message.Method != ""
}

