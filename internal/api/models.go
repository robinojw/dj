package api

import "encoding/json"

// CreateResponseRequest is the request body for POST /v1/responses.
type CreateResponseRequest struct {
	Model              string          `json:"model"`
	Input              json.RawMessage `json:"input"`
	PreviousResponseID string          `json:"previous_response_id,omitempty"`
	Instructions       string          `json:"instructions,omitempty"`
	Tools              []Tool          `json:"tools,omitempty"`
	Reasoning          *Reasoning      `json:"reasoning,omitempty"`
	Stream             bool            `json:"stream"`
}

type Reasoning struct {
	Effort string `json:"effort,omitempty"` // "low", "medium", "high"
}

// Tool represents a tool in the Responses API tools array.
// For function tools, Name/Description/Parameters sit at the top level (not nested).
type Tool struct {
	Type        string          `json:"type"`                  // "function", "mcp", "file_search"
	Name        string          `json:"name,omitempty"`        // function tool name
	Description string          `json:"description,omitempty"` // function tool description
	Parameters  json.RawMessage `json:"parameters,omitempty"`  // function tool parameters JSON schema
	MCP         *MCPTool        `json:"mcp,omitempty"`
}

type MCPTool struct {
	ServerURL    string            `json:"server_url"`
	Headers      map[string]string `json:"headers,omitempty"`
	AllowedTools []string          `json:"allowed_tools,omitempty"`
}

// InputItem represents a single message in the input array.
type InputItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MakeStringInput creates a JSON-encoded string input.
func MakeStringInput(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

// MakeMessagesInput creates a JSON-encoded array of InputItems.
func MakeMessagesInput(items []InputItem) json.RawMessage {
	b, _ := json.Marshal(items)
	return b
}

// --- Response streaming types ---

// ResponseChunk is an event received from the API stream (SSE or WebSocket).
type ResponseChunk struct {
	Type string `json:"type"`

	// For response.output_text.delta
	Delta string `json:"delta,omitempty"`

	// For response.output_item.added — tool calls, text items
	Item *OutputItem `json:"item,omitempty"`

	// For response.completed
	Response *ResponseObject `json:"response,omitempty"`
}

type OutputItem struct {
	Type      string          `json:"type"` // "message", "function_call", "mcp_call"
	ID        string          `json:"id"`
	Name      string          `json:"name,omitempty"`
	Arguments string          `json:"arguments,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type ResponseObject struct {
	ID     string      `json:"id"`
	Status string      `json:"status"`
	Output []OutputItem `json:"output"`
	Usage  Usage       `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// FunctionCallResult is sent back as input to continue after a tool call.
type FunctionCallResult struct {
	Type   string `json:"type"` // "function_call_output"
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

// --- WebSocket event types ---

// wsCreateEvent wraps a CreateResponseRequest for sending over WebSocket.
// The fields are spread at the top level alongside "type": "response.create",
// matching the Responses API WebSocket protocol.
type wsCreateEvent struct {
	Type               string          `json:"type"`
	Model              string          `json:"model"`
	Input              json.RawMessage `json:"input"`
	PreviousResponseID string          `json:"previous_response_id,omitempty"`
	Instructions       string          `json:"instructions,omitempty"`
	Tools              []Tool          `json:"tools,omitempty"`
	Reasoning          *Reasoning      `json:"reasoning,omitempty"`
	Store              *bool           `json:"store,omitempty"`
}

func newWSCreateEvent(req CreateResponseRequest) wsCreateEvent {
	return wsCreateEvent{
		Type:               "response.create",
		Model:              req.Model,
		Input:              req.Input,
		PreviousResponseID: req.PreviousResponseID,
		Instructions:       req.Instructions,
		Tools:              req.Tools,
		Reasoning:          req.Reasoning,
	}
}

// wsErrorEvent represents an error sent by the server over WebSocket.
type wsErrorEvent struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
