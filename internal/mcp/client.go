package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// Client communicates with an MCP server via stdio or HTTP.
type Client struct {
	Config     MCPServerConfig
	serverInfo MCPServerInfo
	tools      []MCPTool

	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Reader
	mu         sync.Mutex
	nextID     atomic.Int64
	httpClient *http.Client
}

func NewClient(cfg MCPServerConfig) *Client {
	return &Client{
		Config:     cfg,
		httpClient: &http.Client{},
	}
}

// Connect establishes the connection and performs the initialize handshake.
func (c *Client) Connect(ctx context.Context) error {
	switch c.Config.Type {
	case "stdio":
		return c.connectStdio(ctx)
	case "http", "sse":
		return c.connectHTTP(ctx)
	default:
		return fmt.Errorf("unsupported MCP server type: %s", c.Config.Type)
	}
}

func (c *Client) connectStdio(ctx context.Context) error {
	parts := strings.Fields(c.Config.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command for MCP server %s", c.Config.Name)
	}

	c.cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start MCP server %s: %w", c.Config.Name, err)
	}

	return c.initialize(ctx)
}

func (c *Client) connectHTTP(ctx context.Context) error {
	return c.initialize(ctx)
}

func (c *Client) initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "dj",
			"version": "0.1.0",
		},
	}

	var result InitializeResult
	if err := c.call(ctx, "initialize", params, &result); err != nil {
		return fmt.Errorf("initialize handshake failed: %w", err)
	}

	c.serverInfo = result.ServerInfo

	_ = c.notify(ctx, "notifications/initialized", nil)

	return nil
}

// ListTools fetches available tools from the MCP server.
func (c *Client) ListTools(ctx context.Context) ([]MCPTool, error) {
	var result ListToolsResult
	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return nil, err
	}
	c.tools = result.Tools
	return result.Tools, nil
}

// CallTool invokes a tool on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*CallToolResult, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": json.RawMessage(args),
	}

	var result CallToolResult
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CachedTools returns tools from the last ListTools call.
func (c *Client) CachedTools() []MCPTool {
	return c.tools
}

// Close shuts down the MCP server connection.
func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	id := int(c.nextID.Add(1))

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	switch c.Config.Type {
	case "stdio":
		return c.callStdio(req, result)
	case "http", "sse":
		return c.callHTTP(ctx, req, result)
	default:
		return fmt.Errorf("unsupported transport: %s", c.Config.Type)
	}
}

func (c *Client) notify(ctx context.Context, method string, params interface{}) error {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	if c.Config.Type == "stdio" {
		c.mu.Lock()
		defer c.mu.Unlock()
		data, err := json.Marshal(req)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c.stdin, "%s\n", data)
		return err
	}
	return nil
}

func (c *Client) callStdio(req jsonRPCRequest, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	if _, err := fmt.Fprintf(c.stdin, "%s\n", data); err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}

	line, err := c.stdout.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read from stdout: %w", err)
	}

	var resp jsonRPCResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		return json.Unmarshal(resp.Result, result)
	}

	return nil
}

func (c *Client) callHTTP(ctx context.Context, req jsonRPCRequest, result interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Config.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range c.Config.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var rpcResp jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result != nil && rpcResp.Result != nil {
		return json.Unmarshal(rpcResp.Result, result)
	}

	return nil
}
