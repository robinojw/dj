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

const jsonRPCVersion = "2.0"

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

	OnEvent  func(message JSONRPCMessage)
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
func (client *Client) Start(ctx context.Context) error {
	client.cmd = exec.CommandContext(ctx, client.command, client.args...)

	var err error
	if err = client.setupPipes(); err != nil {
		return err
	}

	client.scanner = bufio.NewScanner(client.stdout)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)

	if err = client.cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	client.running.Store(true)
	go client.drainStderr()
	return nil
}

func (client *Client) setupPipes() error {
	var err error
	client.stdin, err = client.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	client.stdout, err = client.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	client.stderr, err = client.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	return nil
}

// Running returns true if the child process is alive.
func (client *Client) Running() bool {
	return client.running.Load()
}

// Stop terminates the child process gracefully.
func (client *Client) Stop() error {
	if !client.running.Load() {
		return nil
	}
	client.running.Store(false)

	if client.stdin != nil {
		client.stdin.Close()
	}

	hasProcess := client.cmd != nil && client.cmd.Process != nil
	if hasProcess {
		return client.cmd.Wait()
	}
	return nil
}

func (client *Client) drainStderr() {
	if client.stderr == nil {
		return
	}
	scanner := bufio.NewScanner(client.stderr)
	scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	for scanner.Scan() {
		if client.OnStderr != nil {
			client.OnStderr(scanner.Text())
		}
	}
}

// NextID returns a unique string ID for submissions.
func (client *Client) NextID() string {
	return fmt.Sprintf("dj-%d", client.nextID.Add(1))
}

// ReadLoop reads JSONL messages from stdout and dispatches each to the handler.
func (client *Client) ReadLoop(handler func(JSONRPCMessage)) {
	for client.scanner.Scan() {
		line := client.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var message JSONRPCMessage
		if err := json.Unmarshal(line, &message); err != nil {
			continue
		}

		handler(message)
	}
}
