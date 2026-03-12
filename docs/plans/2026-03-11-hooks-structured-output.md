# Hooks Structured Output Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Return structured output from hooks, add async `on_session_end`, and wire hooks into the worker lifecycle.

**Architecture:** Enhance `Fire()` to return a `HookResult` struct with separated stdout/stderr/exit code. Add `FireAsync()` for fire-and-forget `on_session_end`. Thread the runner from `main.go` through the orchestrator into each worker, firing hooks at tool call boundaries.

**Tech Stack:** Go, `os/exec`, `io`, `bytes`, `time`, bubbletea

---

### Task 1: HookResult Struct and Updated Fire() Signature

**Files:**
- Modify: `internal/hooks/runner.go:1-65`
- Test: `internal/hooks/runner_test.go`

**Step 1: Write failing tests for the new Fire() return type**

Add these tests to `internal/hooks/runner_test.go`:

```go
func TestFireReturnsResult(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "echo hello",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire() error = %v", err)
	}
	if result == nil {
		t.Fatal("Fire() returned nil result for configured hook")
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if result.Stderr != "" {
		t.Errorf("Stderr = %q, want empty", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
	if result.Event != HookPreToolCall {
		t.Errorf("Event = %q, want %q", result.Event, HookPreToolCall)
	}
}

func TestFireCapturesStderr(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookOnError): "echo err >&2",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Fatalf("Fire() error = %v", err)
	}
	if result.Stdout != "" {
		t.Errorf("Stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "err\n" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "err\n")
	}
}

func TestFireNonZeroExit(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "exit 42",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire() infrastructure error = %v", err)
	}
	if result == nil {
		t.Fatal("Fire() returned nil result")
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/robin.white/dev/dj && go test ./internal/hooks/ -run "TestFireReturnsResult|TestFireCapturesStderr|TestFireNonZeroExit" -v`
Expected: FAIL — `Fire()` returns `error`, not `(*HookResult, error)`

**Step 3: Implement HookResult and update Fire()**

Replace `internal/hooks/runner.go` entirely with:

```go
package hooks

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const defaultTimeout = 10 * time.Second

// HookEvent identifies a lifecycle point.
type HookEvent string

const (
	HookPreToolCall  HookEvent = "pre_tool_call"
	HookPostToolCall HookEvent = "post_tool_call"
	HookOnError      HookEvent = "on_error"
	HookSessionEnd   HookEvent = "on_session_end"
)

// HookResult captures the outcome of a hook execution.
type HookResult struct {
	Event    HookEvent
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error // non-nil only for infrastructure failures (sh not found, timeout)
}

// Config holds hook shell command templates.
type Config struct {
	Hooks   map[string]string
	Timeout time.Duration // 0 means use defaultTimeout
}

func (c Config) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultTimeout
}

// Runner executes configured hooks at lifecycle points.
type Runner struct {
	config Config
}

func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Fire executes the hook for the given event with variable substitution.
// Returns (nil, nil) if no hook is configured for the event.
// Returns (*HookResult, nil) on both success and non-zero exit.
// The error return is reserved for infrastructure failures.
func (r *Runner) Fire(event HookEvent, vars map[string]string) (*HookResult, error) {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return nil, nil
	}

	expanded := expandVars(cmdTemplate, vars)

	cmd := exec.Command("sh", "-c", expanded)
	cmd.WaitDelay = r.config.timeout()

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &HookResult{
		Event:    event,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			// Infrastructure failure (e.g., sh not found, timeout)
			result.Err = fmt.Errorf("hook %s failed: %w", event, err)
			return result, result.Err
		}
	}

	return result, nil
}

// FireAsync launches the hook in a background goroutine (fire-and-forget).
// Used for on_session_end where blocking is undesirable.
func (r *Runner) FireAsync(event HookEvent, vars map[string]string) {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return
	}

	expanded := expandVars(cmdTemplate, vars)

	go func() {
		cmd := exec.Command("sh", "-c", expanded)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		cmd.WaitDelay = r.config.timeout()
		_ = cmd.Run()
	}()
}

// expandVars replaces {{key}} placeholders with values from vars.
func expandVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
```

**Step 4: Fix existing tests that use the old Fire() signature**

Update `TestFireHook` in `internal/hooks/runner_test.go` — change `err := runner.Fire(...)` to `_, err := runner.Fire(...)`.

Update `TestFireUnconfiguredHook` — change to:

```go
func TestFireUnconfiguredReturnsNil(t *testing.T) {
	cfg := Config{Hooks: map[string]string{}}
	runner := NewRunner(cfg)

	result, err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result for unconfigured hook, got %+v", result)
	}
}
```

**Step 5: Run all hooks tests to verify they pass**

Run: `cd /Users/robin.white/dev/dj && go test ./internal/hooks/ -v`
Expected: PASS (all tests including new ones)

**Step 6: Commit**

```bash
git add internal/hooks/runner.go internal/hooks/runner_test.go
git commit -m "feat(hooks): return structured HookResult from Fire() (#17)

Add HookResult struct with separated stdout/stderr, exit code, and
duration. Fire() now returns (*HookResult, nil) for both success and
non-zero exits; error is reserved for infrastructure failures."
```

---

### Task 2: FireAsync and Timeout Tests

**Files:**
- Test: `internal/hooks/runner_test.go`

**Step 1: Write failing tests for FireAsync and timeout**

Add to `internal/hooks/runner_test.go`:

```go
func TestFireAsyncReturnsImmediately(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookSessionEnd): "sleep 5",
		},
	}
	runner := NewRunner(cfg)

	start := time.Now()
	runner.FireAsync(HookSessionEnd, nil)
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("FireAsync blocked for %v, expected immediate return", elapsed)
	}
}

func TestFireTimeout(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "sleep 10",
		},
		Timeout: 1 * time.Second,
	}
	runner := NewRunner(cfg)

	start := time.Now()
	result, err := runner.Fire(HookPreToolCall, nil)
	elapsed := time.Since(start)

	// Should complete within ~2s (1s timeout + WaitDelay grace)
	if elapsed > 5*time.Second {
		t.Errorf("Fire() took %v, expected timeout around 1-2s", elapsed)
	}

	// Either err is non-nil (infrastructure failure) or result has non-zero exit
	if err == nil && result != nil && result.ExitCode == 0 && result.Err == nil {
		t.Error("Expected timeout to produce an error or non-zero exit")
	}
}
```

**Step 2: Run tests to verify they pass**

Run: `cd /Users/robin.white/dev/dj && go test ./internal/hooks/ -run "TestFireAsync|TestFireTimeout" -v -timeout 30s`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/hooks/runner_test.go
git commit -m "test(hooks): add FireAsync and timeout tests (#17)"
```

---

### Task 3: Add UpdateHookResult Message Type

**Files:**
- Modify: `internal/agents/messages.go:17-24`

**Step 1: Add UpdateHookResult and HookResult field**

In `internal/agents/messages.go`, add `UpdateHookResult` to the `UpdateType` const block (after `UpdateError`) and add a `HookResult` field to `WorkerUpdate`. The import for hooks is needed.

Change the imports to:

```go
import (
	"time"

	"github.com/robinojw/dj/internal/hooks"
)
```

Add after `UpdateError`:

```go
	UpdateHookResult                   // hook execution result
```

Add to the `WorkerUpdate` struct after `DiffInfo`:

```go
	HookResult *hooks.HookResult // populated when Type == UpdateHookResult
```

**Step 2: Verify compilation**

Run: `cd /Users/robin.white/dev/dj && go build ./internal/agents/`
Expected: Compiles without errors

**Step 3: Commit**

```bash
git add internal/agents/messages.go
git commit -m "feat(agents): add UpdateHookResult message type (#17)"
```

---

### Task 4: Add Timeout to HooksConfig

**Files:**
- Modify: `config/config.go:34-39`

**Step 1: Add Timeout field to HooksConfig**

In `config/config.go`, add the `Timeout` field:

```go
type HooksConfig struct {
	Timeout      string `toml:"timeout"`
	PreToolCall  string `toml:"pre_tool_call"`
	PostToolCall string `toml:"post_tool_call"`
	OnError      string `toml:"on_error"`
	OnSessionEnd string `toml:"on_session_end"`
}
```

**Step 2: Add commented hooks section to harness.toml**

Append to the end of `harness.toml`:

```toml
# [hooks]
# timeout = "10s"
# pre_tool_call = "echo 'Tool: {{tool_name}}'"
# post_tool_call = ""
# on_error = ""
# on_session_end = ""
```

**Step 3: Verify compilation**

Run: `cd /Users/robin.white/dev/dj && go build ./config/`
Expected: Compiles without errors

**Step 4: Commit**

```bash
git add config/config.go harness.toml
git commit -m "feat(config): add hooks timeout setting and documented example (#17)"
```

---

### Task 5: Wire Hooks into Worker

**Files:**
- Modify: `internal/agents/worker.go:18-59` (struct and constructor)
- Modify: `internal/agents/worker.go:102-115` (pre_tool_call)
- Modify: `internal/agents/worker.go:117-131` (post_tool_call)
- Modify: `internal/agents/worker.go:80-90` (on_error — context cancel path)
- Modify: `internal/agents/worker.go:148-157` (on_error — stream error path)
- Test: `internal/agents/worker_test.go` (new file or append to existing)

**Step 1: Write failing test for nil hooks safety**

Create or append to `internal/agents/worker_test.go`:

```go
func TestWorkerNilHooksDoesNotPanic(t *testing.T) {
	// Constructing a worker with nil hooks should not panic
	task := Subtask{ID: "test-1", Description: "test task"}
	w := NewWorker(task, nil, nil, "test-model", "parent", ModeNormal, nil, nil, nil, nil)
	if w.hooks != nil {
		t.Error("Expected nil hooks on worker")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -run TestWorkerNilHooksDoesNotPanic -v`
Expected: FAIL — `NewWorker` doesn't accept a hooks parameter

**Step 3: Add hooks field to Worker and update NewWorker**

In `internal/agents/worker.go`, add import:

```go
	"github.com/robinojw/dj/internal/hooks"
```

Add field to `Worker` struct (after `permReqCh`):

```go
	hooks        *hooks.Runner
```

Update `NewWorker` signature to accept `hooks *hooks.Runner` as the last parameter:

```go
func NewWorker(
	task Subtask,
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	model string,
	parentID string,
	mode AgentMode,
	mem *memory.Manager,
	gate *modes.Gate,
	permReqCh chan<- modes.PermissionRequest,
	hooks *hooks.Runner,
) *Worker {
	return &Worker{
		ID:        task.ID,
		Task:      task,
		Status:    "pending",
		Mode:      mode,
		client:    client,
		skills:    skillsRegistry,
		memory:    mem,
		model:     model,
		parentID:  parentID,
		gate:      gate,
		permReqCh: permReqCh,
		hooks:     hooks,
	}
}
```

**Step 4: Add fireHook helper method**

Add to `internal/agents/worker.go` (after `buildInstructions`):

```go
// fireHook fires a hook and sends the result as an update. Advisory only — never fatal.
func (w *Worker) fireHook(event hooks.HookEvent, vars map[string]string, updates chan<- WorkerUpdate) {
	if w.hooks == nil {
		return
	}
	result, _ := w.hooks.Fire(event, vars)
	if result != nil {
		updates <- WorkerUpdate{
			WorkerID:   w.ID,
			Type:       UpdateHookResult,
			HookResult: result,
		}
	}
}
```

**Step 5: Fire pre_tool_call in the output_item.added case**

In `Worker.Run()`, in the `response.output_item.added` case, after tracking args (line ~108) and before sending `UpdateToolCall` (line ~110), add:

```go
				argsJSON, _ := json.Marshal(w.lastToolArgs)
				w.fireHook(hooks.HookPreToolCall, map[string]string{
					"tool_name": chunk.Item.Name,
					"tool_args": string(argsJSON),
					"worker_id": w.ID,
				}, updates)
```

**Step 6: Fire post_tool_call in the function_call_result case**

In `Worker.Run()`, in the `response.function_call_result` case, before the `isEditTool` check (line ~119), add:

```go
			argsJSON, _ := json.Marshal(w.lastToolArgs)
			w.fireHook(hooks.HookPostToolCall, map[string]string{
				"tool_name": w.lastToolName,
				"tool_args": string(argsJSON),
				"worker_id": w.ID,
			}, updates)
```

**Step 7: Fire on_error in both error paths**

In the context cancellation path (line ~83, inside `case <-ctx.Done():`), before sending `UpdateError`, add:

```go
			w.fireHook(hooks.HookOnError, map[string]string{
				"error_msg": ctx.Err().Error(),
				"worker_id": w.ID,
				"tool_name": w.lastToolName,
			}, updates)
```

In the stream error loop (line ~149, `for err := range errs`), before sending `UpdateError`, add:

```go
		w.fireHook(hooks.HookOnError, map[string]string{
			"error_msg": err.Error(),
			"worker_id": w.ID,
			"tool_name": w.lastToolName,
		}, updates)
```

**Step 8: Fix the Orchestrator's call to NewWorker**

In `internal/agents/orchestrator.go:51`, update the `NewWorker` call to pass `o.Hooks`:

```go
		w := NewWorker(task, o.client, o.skills, o.model, o.RootID, o.Mode, o.Memory, o.Gate, o.PermReqCh, o.Hooks)
```

Add the `Hooks` field to `Orchestrator` struct (after `PermReqCh`):

```go
	Hooks     *hooks.Runner
```

Add the import:

```go
	"github.com/robinojw/dj/internal/hooks"
```

**Step 9: Run all tests to verify they pass**

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -v`
Expected: PASS

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Build will fail because `main.go` and `app.go` still use old signatures. That's expected — we fix those in Task 6.

**Step 10: Commit**

```bash
git add internal/agents/worker.go internal/agents/worker_test.go internal/agents/orchestrator.go internal/agents/messages.go
git commit -m "feat(agents): wire hooks into worker lifecycle (#17)

Worker fires pre_tool_call, post_tool_call, and on_error hooks at the
appropriate points. Results are sent as UpdateHookResult messages.
Nil hooks runner is safe (no-op)."
```

---

### Task 6: Wire Hooks Through main.go and TUI

**Files:**
- Modify: `cmd/harness/main.go:93-105`
- Modify: `internal/tui/app.go:64-98` (NewApp signature)
- Modify: `internal/tui/app.go:219-229` (Update handler)

**Step 1: Update main.go to parse timeout and pass hooks to NewApp**

In `cmd/harness/main.go`, replace lines 93-105 with:

```go
	// Set up event hooks
	var hookTimeout time.Duration
	if cfg.Hooks.Timeout != "" {
		if parsed, err := time.ParseDuration(cfg.Hooks.Timeout); err == nil {
			hookTimeout = parsed
		} else {
			fmt.Fprintf(os.Stderr, "Warning: invalid hooks timeout %q: %v\n", cfg.Hooks.Timeout, err)
		}
	}
	hookRunner := hooks.NewRunner(hooks.Config{
		Hooks: map[string]string{
			string(hooks.HookPreToolCall):  cfg.Hooks.PreToolCall,
			string(hooks.HookPostToolCall): cfg.Hooks.PostToolCall,
			string(hooks.HookOnError):      cfg.Hooks.OnError,
			string(hooks.HookSessionEnd):   cfg.Hooks.OnSessionEnd,
		},
		Timeout: hookTimeout,
	})
	defer hookRunner.FireAsync(hooks.HookSessionEnd, map[string]string{"summary": "session ended"})

	app := tui.NewApp(t, client, tracker, cfg.Model.Default, cfg, hookRunner)
```

Add `"time"` to the imports in `main.go` if not already present.

**Step 2: Update NewApp to accept hooks runner**

In `internal/tui/app.go`, change the `NewApp` signature:

```go
func NewApp(
	t *theme.Theme,
	client *api.ResponsesClient,
	tracker *api.Tracker,
	model string,
	cfg config.Config,
	hookRunner *hooks.Runner,
) App {
```

In the `App` literal inside `NewApp`, add after `debugOverlay`:

```go
		hooks:           hookRunner,
```

**Step 3: Handle UpdateHookResult in App.Update()**

In `internal/tui/app.go`, in the `agents.WorkerUpdate` case (line ~219), add a handler for hook results before the existing diff handler:

```go
	case agents.WorkerUpdate:
		// Log hook results to debug overlay
		if msg.Type == agents.UpdateHookResult && msg.HookResult != nil {
			if a.debugMode {
				info := fmt.Sprintf("Hook %s: exit=%d stdout=%q",
					msg.HookResult.Event, msg.HookResult.ExitCode, msg.HookResult.Stdout)
				a.debugOverlay.AddInfo(info)
			}
			return a, nil
		}

		// Convert UpdateDiffResult to StreamDiffMsg for the UI
		if msg.Type == agents.UpdateDiffResult && msg.DiffInfo != nil {
```

**Step 4: Verify full build**

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Compiles without errors

**Step 5: Run all tests**

Run: `cd /Users/robin.white/dev/dj && go test ./...`
Expected: PASS

**Step 6: Commit**

```bash
git add cmd/harness/main.go internal/tui/app.go
git commit -m "feat(tui): wire hooks through main.go and TUI app (#17)

Parse hooks timeout from config, pass runner to NewApp, use FireAsync
for on_session_end, and route UpdateHookResult to debug overlay."
```

---

### Task 7: Final Verification and Cleanup

**Files:**
- All modified files

**Step 1: Run full test suite**

Run: `cd /Users/robin.white/dev/dj && go test ./... -v`
Expected: All PASS

**Step 2: Run vet and build**

Run: `cd /Users/robin.white/dev/dj && go vet ./... && go build ./...`
Expected: No issues

**Step 3: Verify no leftover placeholder comments**

Search for `// will be wired` in the codebase — the `_ = hookRunner` line in `main.go` should be gone.

Run: `grep -r "will be wired" cmd/ internal/`
Expected: Only `_ = lspClient` and `_ = memMgr` remain (unrelated to hooks)

**Step 4: Commit any cleanup if needed**

If all clean, no commit needed. Otherwise:

```bash
git commit -am "chore: clean up hook wiring placeholders (#17)"
```
