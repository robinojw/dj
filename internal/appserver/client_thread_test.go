package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestClientCreateThread(t *testing.T) {
	clientRead, serverWrite := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	go mockThreadCreateServer(t, serverRead, serverWrite)

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	client.running.Store(true)

	go client.ReadLoop(client.Dispatch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.CreateThread(ctx, "Build a web app")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}
	if result.ThreadID != "t-new-123" {
		t.Errorf("expected t-new-123, got %s", result.ThreadID)
	}
}

func mockThreadCreateServer(t *testing.T, reader *io.PipeReader, writer *io.PipeWriter) {
	t.Helper()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		t.Error("mock: failed to read request")
		return
	}
	var req Message
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		t.Errorf("mock: unmarshal: %v", err)
		return
	}

	resp := Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  json.RawMessage(`{"threadId":"t-new-123"}`),
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	writer.Write(data)
}

func TestClientListThreads(t *testing.T) {
	clientRead, serverWrite := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	go mockThreadListServer(t, serverRead, serverWrite)

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	client.running.Store(true)

	go client.ReadLoop(client.Dispatch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.ListThreads(ctx)
	if err != nil {
		t.Fatalf("ListThreads failed: %v", err)
	}
	if len(result.Threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(result.Threads))
	}
}

func mockThreadListServer(t *testing.T, reader *io.PipeReader, writer *io.PipeWriter) {
	t.Helper()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		t.Error("mock: failed to read request")
		return
	}
	var req Message
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		t.Errorf("mock: unmarshal: %v", err)
		return
	}

	threadList := `{"threads":[{"id":"t-1","status":"active","title":"A"},{"id":"t-2","status":"idle","title":"B"}]}`
	resp := Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  json.RawMessage(threadList),
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	writer.Write(data)
}
