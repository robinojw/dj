# Phase 1: Project Scaffold + App-Server IPC Client

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Initialize the Go project and build a bidirectional JSON-RPC 2.0 client that spawns the Codex App Server, sends requests, and receives notifications over stdio.

**Architecture:** The client spawns `codex app-server --listen stdio://` as a child process. A dedicated goroutine reads newline-delimited JSON from stdout and dispatches messages. Requests are written to stdin as JSONL with a mutex. A pending-request map tracks `id → response channel` for synchronous-style calls.

**Tech Stack:** Go 1.22+, `os/exec`, `encoding/json`, `bufio`, `sync`

**Status:** COMPLETE

---

### Task 1: Initialize Go Module and Directory Structure
- Created `go.mod`, `cmd/dj/main.go` entry point
- Commit: `feat: initialize go module and project scaffold`

### Task 2: Define JSON-RPC 2.0 Base Types
- `internal/appserver/protocol.go`: Message, Request, Response, RPCError
- 4 tests: marshal/unmarshal for requests, responses, notifications, errors
- Commit: `feat(appserver): define JSON-RPC 2.0 base types`

### Task 3: Build the IPC Client — Process Lifecycle
- `internal/appserver/client.go`: Client struct, NewClient, Start, Stop, Running
- 1 test using `cat` as mock process
- Commit: `feat(appserver): client process lifecycle start/stop`

### Task 4: Client Send and ReadLoop
- Send() writes JSONL to stdin, ReadLoop() reads JSONL from stdout
- 1 test: echo round-trip via `cat`
- Commit: `feat(appserver): send requests and read JSONL responses`

### Task 5: Synchronous Call with Pending Request Tracking
- Call() sends request and blocks until matching response via channel
- Dispatch() routes messages to pending calls, server requests, or notifications
- 1 test: Call + Dispatch via `cat` echo
- Commit: `feat(appserver): synchronous Call with pending request tracking`

### Task 6: Initialize Handshake
- Initialize() sends `initialize` request, receives capabilities, sends `initialized` notification
- 1 test: mock server with bidirectional io.Pipe
- Commit: `feat(appserver): initialize/initialized handshake`

### Task 7: Integration Smoke Test
- `integration_test.go` with `//go:build integration` tag
- Tests real `codex app-server` connection when available
- Commit: `test(appserver): integration smoke test with real app-server`
