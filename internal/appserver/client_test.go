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

	msgs := make(chan Message, 10)
	go client.ReadLoop(func(msg Message) {
		msgs <- msg
	})

	req := &Request{
		ID:     intPtr(1),
		Method: "test/echo",
		Params: json.RawMessage(`{"hello":"world"}`),
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

func TestClientCall(t *testing.T) {
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	go client.ReadLoop(client.Dispatch)

	resp, err := client.Call(ctx, "test/method", json.RawMessage(`{"key":"val"}`))
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestInitializeHandshake(t *testing.T) {
	clientRead, serverWrite := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(serverRead)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		if !scanner.Scan() {
			t.Error("mock server: failed to read initialize request")
			return
		}
		var req Message
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			t.Errorf("mock server: unmarshal request: %v", err)
			return
		}
		if req.Method != "initialize" {
			t.Errorf("mock server: expected method initialize, got %s", req.Method)
			return
		}

		resp := Message{
			ID:     req.ID,
			Result: json.RawMessage(`{"serverInfo":{"name":"codex-app-server","version":"0.1.0"}}`),
		}
		data, _ := json.Marshal(resp)
		data = append(data, '\n')
		serverWrite.Write(data)

		if !scanner.Scan() {
			t.Error("mock server: failed to read initialized notification")
			return
		}
		var notif Message
		if err := json.Unmarshal(scanner.Bytes(), &notif); err != nil {
			t.Errorf("mock server: unmarshal notification: %v", err)
			return
		}
		if notif.Method != "initialized" {
			t.Errorf("mock server: expected method initialized, got %s", notif.Method)
		}
		if notif.Params == nil {
			t.Error("mock server: initialized notification must include params")
		}
	}()

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	client.running.Store(true)

	go client.ReadLoop(client.Dispatch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	caps, err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if caps == nil {
		t.Fatal("expected non-nil capabilities")
	}
	if caps.ServerInfo.Name != "codex-app-server" {
		t.Errorf("expected server name codex-app-server, got %s", caps.ServerInfo.Name)
	}
	if caps.ServerInfo.Version != "0.1.0" {
		t.Errorf("expected server version 0.1.0, got %s", caps.ServerInfo.Version)
	}
}
