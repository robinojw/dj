package mcp

import "encoding/json"

// MCPServerConfig holds the configuration for a single MCP server.
type MCPServerConfig struct {
	Name      string            `toml:"name"`
	Type      string            `toml:"type"`    // "stdio", "http", "sse"
	Command   string            `toml:"command"` // for stdio
	URL       string            `toml:"url"`     // for http/sse
	Headers   map[string]string `toml:"headers"`
	AutoStart bool              `toml:"auto_start"`
}

// MCPTool represents a tool exposed by an MCP server.
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// MCPResource represents a resource exposed by an MCP server.
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// MCPPrompt represents a prompt template exposed by an MCP server.
type MCPPrompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []MCPPromptArg   `json:"arguments"`
}

type MCPPromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// --- JSON-RPC 2.0 types ---

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeResult is the response from the MCP initialize handshake.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      MCPServerInfo `json:"serverInfo"`
	Capabilities    interface{}  `json:"capabilities"`
}

type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListToolsResult is the response from tools/list.
type ListToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

// CallToolResult is the response from tools/call.
type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
