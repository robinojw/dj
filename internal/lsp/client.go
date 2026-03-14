package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const diagnosticCollectTimeout = 500 * time.Millisecond

// Client manages a running LSP server process.
type Client struct {
	config   ServerConfig
	rootPath string
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	mu       sync.Mutex
	nextID   atomic.Int64
	diagMu   sync.Mutex
	diags    map[string][]Diagnostic
}

// NewClient creates a new LSP client but does not start the server.
func NewClient(cfg ServerConfig, rootPath string) *Client {
	return &Client{
		config:   cfg,
		rootPath: rootPath,
		diags:    make(map[string][]Diagnostic),
	}
}

// Start launches the LSP server and performs the initialize handshake.
func (c *Client) Start(ctx context.Context) error {
	args := c.config.Args
	c.cmd = exec.CommandContext(ctx, c.config.Command, args...)

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
		return fmt.Errorf("start LSP server %s: %w", c.config.Command, err)
	}

	return c.initialize(ctx)
}

func (c *Client) initialize(_ context.Context) error {
	params := map[string]interface{}{
		"processId": nil,
		"rootUri":   "file://" + c.rootPath,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"publishDiagnostics": map[string]interface{}{},
			},
		},
	}

	var result json.RawMessage
	if err := c.call("initialize", params, &result); err != nil {
		return fmt.Errorf("LSP initialize: %w", err)
	}

	return c.notify("initialized", struct{}{})
}

// NotifyChange sends textDocument/didOpen and collects diagnostics.
func (c *Client) NotifyChange(file string, content string) ([]Diagnostic, error) {
	uri := "file://" + file

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": c.config.Language,
			"version":    1,
			"text":       content,
		},
	}
	if err := c.notify("textDocument/didOpen", params); err != nil {
		return nil, err
	}

	return c.collectDiagnostics(file, diagnosticCollectTimeout), nil
}

func (c *Client) collectDiagnostics(file string, timeout time.Duration) []Diagnostic {
	time.Sleep(timeout)
	c.diagMu.Lock()
	defer c.diagMu.Unlock()
	return c.diags[file]
}

// Language returns the language this client handles.
func (c *Client) Language() string {
	return c.config.Language
}

// Command returns the LSP server command name.
func (c *Client) Command() string {
	return c.config.Command
}

// Close shuts down the LSP server.
func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// FormatDiagnostic returns a human-readable string for a diagnostic.
func FormatDiagnostic(d Diagnostic) string {
	return fmt.Sprintf("%s:%d:%d: %s: %s (%s)",
		d.File, d.Line, d.Column, d.Severity, d.Message, d.Source)
}

// FormatDiagnostics returns a newline-separated string of all diagnostics.
func FormatDiagnostics(diags []Diagnostic) string {
	lines := make([]string, len(diags))
	for i, d := range diags {
		lines[i] = FormatDiagnostic(d)
	}
	return strings.Join(lines, "\n")
}

func (c *Client) call(method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := int(c.nextID.Add(1))

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	if err := c.writeMessage(req); err != nil {
		return err
	}

	return c.readResponse(result)
}

func (c *Client) notify(method string, params interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	return c.writeMessage(req)
}

func (c *Client) writeMessage(msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	_, err = fmt.Fprintf(c.stdin, "%s%s", header, body)
	return err
}

func (c *Client) readResponse(result interface{}) error {
	var contentLength int
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}

	if contentLength == 0 {
		return fmt.Errorf("no Content-Length in response")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, body); err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("LSP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		return json.Unmarshal(resp.Result, result)
	}

	return nil
}
