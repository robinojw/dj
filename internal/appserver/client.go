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

// Client manages a child codex proto process and bidirectional JSONL communication.
type Client struct {
	command string
	args    []string

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	mu     sync.Mutex
	nextID atomic.Int64

	running atomic.Bool

	Router *EventRouter
}

const scannerBufferSize = 1024 * 1024

// NewClient creates a client that will spawn the given command.
func NewClient(command string, args ...string) *Client {
	return &Client{
		command: command,
		args:    args,
	}
}

// Start spawns the child process.
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
	c.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)

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

	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Wait()
	}
	return nil
}

// Send writes a Submission to the child's stdin as a JSONL line.
func (c *Client) Send(sub *Submission) error {
	data, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("marshal submission: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	data = append(data, '\n')
	_, err = c.stdin.Write(data)
	return err
}

// NextID generates a unique string ID for a submission.
func (c *Client) NextID() string {
	return fmt.Sprintf("sub-%d", c.nextID.Add(1))
}

// ReadLoop reads JSONL events from stdout and dispatches to the router.
// Blocks until stdout is closed or an error occurs.
func (c *Client) ReadLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		if c.Router != nil {
			c.Router.HandleEvent(event)
		}
	}
	c.running.Store(false)
}
