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

const scannerBufferSize = 1024 * 1024

// Client manages a child codex proto process and bidirectional communication.
type Client struct {
	command string
	args    []string

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	scanner *bufio.Scanner

	mu      sync.Mutex
	nextID  atomic.Int64
	running atomic.Bool

	OnEvent  func(event ProtoEvent)
	OnStderr func(line string)
}

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

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	c.scanner = bufio.NewScanner(c.stdout)
	c.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	c.running.Store(true)
	go c.drainStderr()
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

func (c *Client) drainStderr() {
	if c.stderr == nil {
		return
	}
	scanner := bufio.NewScanner(c.stderr)
	scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	for scanner.Scan() {
		if c.OnStderr != nil {
			c.OnStderr(scanner.Text())
		}
	}
}

// Send writes a ProtoSubmission to the child's stdin as a JSONL line.
func (c *Client) Send(sub *ProtoSubmission) error {
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

// NextID returns a unique string ID for submissions.
func (c *Client) NextID() string {
	return fmt.Sprintf("dj-%d", c.nextID.Add(1))
}

// ReadLoop reads JSONL events from stdout and dispatches each to the handler.
func (c *Client) ReadLoop(handler func(ProtoEvent)) {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event ProtoEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		handler(event)
	}
}

// SendUserInput sends a text message to the codex session.
func (c *Client) SendUserInput(text string) (string, error) {
	id := c.NextID()
	op := UserInputOp{
		Type: OpUserInput,
		Items: []InputItem{
			{Type: "text", Text: text},
		},
	}
	opData, _ := json.Marshal(op)
	sub := &ProtoSubmission{
		ID: id,
		Op: opData,
	}
	return id, c.Send(sub)
}

// SendInterrupt sends an interrupt to cancel the current task.
func (c *Client) SendInterrupt() error {
	id := c.NextID()
	op := map[string]string{"type": OpInterrupt}
	opData, _ := json.Marshal(op)
	return c.Send(&ProtoSubmission{ID: id, Op: opData})
}

// SendApproval responds to an exec or patch approval request.
func (c *Client) SendApproval(eventID string, opType string, approved bool) error {
	op := map[string]any{"type": opType, "approved": approved}
	opData, _ := json.Marshal(op)
	return c.Send(&ProtoSubmission{ID: eventID, Op: opData})
}
