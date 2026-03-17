package appserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestClientStartStop(t *testing.T) {
	// Use 'cat' as a mock app-server: it reads stdin and echoes to stdout.
	// This verifies process lifecycle without a real codex binary.
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !client.Running() {
		t.Fatal("expected client to be running")
	}

	if err := client.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if client.Running() {
		t.Fatal("expected client to be stopped")
	}
}

func TestClientSendAndRead(t *testing.T) {
	// 'cat' echoes back what we write — simulates a response
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	// Start the read loop
	msgs := make(chan Message, 10)
	go client.ReadLoop(func(msg Message) {
		msgs <- msg
	})

	// Send a JSON-RPC request — cat will echo it back
	req := &Request{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "test/echo",
		Params:  json.RawMessage(`{"hello":"world"}`),
	}
	if err := client.Send(req); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case msg := <-msgs:
		if msg.Method != "test/echo" {
			t.Errorf("expected method test/echo, got %s", msg.Method)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}
