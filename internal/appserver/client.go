package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client manages a child app-server process and bidirectional JSON-RPC communication.
type Client struct {
	command string
	args    []string

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	mu      sync.Mutex // protects writes to stdin
	nextID  atomic.Int64
	pending sync.Map // id → chan *Message

	running atomic.Bool

	// OnNotification is called for each server notification (no id).
	// Set this before calling Start.
	OnNotification func(method string, params json.RawMessage)

	// OnServerRequest is called for server-to-client requests (has id).
	// Set this before calling Start.
	OnServerRequest func(id int, method string, params json.RawMessage)

	// Router dispatches typed notifications by method name.
	// Falls back to OnNotification for unregistered methods.
	Router *NotificationRouter
}

// NewClient creates a client that will spawn the given command.
// Additional arguments can be passed after the command.
func NewClient(command string, args ...string) *Client {
	return &Client{
		command: command,
		args:    args,
	}
}

// Start spawns the child process and begins reading stdout.
func (c *Client) Start(ctx context.Context) error {
	c.cmd = exec.CommandContext(ctx, c.command, c.args...)

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	c.scanner = bufio.NewScanner(c.stdout)
	c.scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max line

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	c.running.Store(true)
	return nil
}

// Running returns true if the child process is alive.
func (c *Client) Running() bool {
	return c.running.Load()
}

// Stop terminates the child process gracefully.
func (c *Client) Stop() error {
	if !c.running.Load() {
		return nil
	}
	c.running.Store(false)

	// Close stdin to signal EOF to the child
	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Wait()
	}
	return nil
}

// Send writes a JSON-RPC request to the child's stdin as a JSONL line.
func (c *Client) Send(req *Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	data = append(data, '\n')
	_, err = c.stdin.Write(data)
	return err
}

// ReadLoop reads JSONL from stdout and dispatches each message to the callback.
// It blocks until the scanner is exhausted (stdout closed) or an error occurs.
func (c *Client) ReadLoop(handler func(Message)) {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			continue // skip malformed lines
		}

		handler(msg)
	}
}

// Call sends a request and blocks until the response with the matching ID arrives.
func (c *Client) Call(ctx context.Context, method string, params json.RawMessage) (*Message, error) {
	id := int(c.nextID.Add(1))

	ch := make(chan *Message, 1)
	c.pending.Store(id, ch)
	defer c.pending.Delete(id)

	req := &Request{
		ID:     &id,
		Method: method,
		Params: params,
	}

	if err := c.Send(req); err != nil {
		return nil, err
	}

	select {
	case msg := <-ch:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Dispatch routes an incoming message to the appropriate handler:
// - Messages with an ID matching a pending request -> resolve the pending Call
// - Messages with an ID but no pending request -> server-to-client request (OnServerRequest)
// - Messages without an ID -> notification (OnNotification)
func (c *Client) Dispatch(msg Message) {
	if msg.ID != nil {
		// Check if this resolves a pending call
		if ch, ok := c.pending.LoadAndDelete(*msg.ID); ok {
			ch.(chan *Message) <- &msg
			return
		}

		// Server-to-client request
		if c.OnServerRequest != nil && msg.Method != "" {
			c.OnServerRequest(*msg.ID, msg.Method, msg.Params)
		}
		return
	}

	if msg.Method == "" {
		return
	}

	if c.Router != nil {
		c.Router.Handle(msg.Method, msg.Params)
	}

	if c.OnNotification != nil {
		c.OnNotification(msg.Method, msg.Params)
	}
}

// InitializeParams is sent as the first request to the app-server.
type InitializeParams struct {
	ClientInfo ClientInfo `json:"clientInfo"`
}

// ClientInfo identifies this client to the app-server.
type ClientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

// ServerCapabilities is the result of the initialize request.
type ServerCapabilities struct {
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// Initialize performs the required handshake with the app-server.
// Sends initialize request, receives capabilities, then sends initialized notification.
func (c *Client) Initialize(ctx context.Context) (*ServerCapabilities, error) {
	params, _ := json.Marshal(InitializeParams{
		ClientInfo: ClientInfo{
			Name:    "dj",
			Title:   "DJ — Codex TUI Visualizer",
			Version: "0.1.0",
		},
	})

	resp, err := c.Call(ctx, "initialize", params)
	if err != nil {
		return nil, fmt.Errorf("initialize request: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	var caps ServerCapabilities
	if resp.Result != nil {
		if err := json.Unmarshal(resp.Result, &caps); err != nil {
			return nil, fmt.Errorf("unmarshal capabilities: %w", err)
		}
	}

	// Send the initialized notification (no id, no response expected).
	// Codex requires params: {} even for empty notifications.
	notif := &Request{
		Method: "initialized",
		Params: json.RawMessage(`{}`),
	}
	if err := c.Send(notif); err != nil {
		return nil, fmt.Errorf("send initialized: %w", err)
	}

	return &caps, nil
}
