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
