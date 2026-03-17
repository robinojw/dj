package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestClientStartStop(t *testing.T) {
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
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	events := make(chan Event, 10)
	client.Router = NewEventRouter()
	client.Router.OnSessionConfigured(func(event SessionConfigured) {
		events <- Event{Msg: json.RawMessage(`{"type":"session_configured"}`)}
	})

	go client.ReadLoop()

	sub := &Submission{
		ID: "test-1",
		Op: json.RawMessage(`{"type":"session_configured","session_id":"s1","model":"test"}`),
	}

	wrappedEvent := Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"session_configured","session_id":"s1","model":"test"}`),
	}
	data, _ := json.Marshal(wrappedEvent)
	client.mu.Lock()
	data = append(data, '\n')
	client.stdin.Write(data)
	client.mu.Unlock()
	_ = sub

	select {
	case <-events:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestClientNextID(t *testing.T) {
	client := NewClient("cat")

	id1 := client.NextID()
	id2 := client.NextID()

	if id1 == id2 {
		t.Errorf("expected unique IDs, got %s and %s", id1, id2)
	}
	if id1 != "sub-1" {
		t.Errorf("expected sub-1, got %s", id1)
	}
	if id2 != "sub-2" {
		t.Errorf("expected sub-2, got %s", id2)
	}
}

func TestClientReadLoopParsesEvents(t *testing.T) {
	clientRead, serverWrite := io.Pipe()

	client := &Client{}
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	received := make(chan SessionConfigured, 1)
	client.Router = NewEventRouter()
	client.Router.OnSessionConfigured(func(event SessionConfigured) {
		received <- event
	})

	go client.ReadLoop()

	eventJSON := `{"id":"","msg":{"type":"session_configured","session_id":"sess-123","model":"gpt-4o"}}` + "\n"
	serverWrite.Write([]byte(eventJSON))

	select {
	case event := <-received:
		if event.SessionID != "sess-123" {
			t.Errorf("expected sess-123, got %s", event.SessionID)
		}
		if event.Model != "gpt-4o" {
			t.Errorf("expected gpt-4o, got %s", event.Model)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for session_configured event")
	}

	serverWrite.Close()
}
