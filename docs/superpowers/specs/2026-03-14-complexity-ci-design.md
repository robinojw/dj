# Complexity CI Checks Design

## Overview

Add code complexity enforcement to CI using `golangci-lint` with `funlen` and `cyclop` linters. Refactor existing code that exceeds thresholds.

## Thresholds

| Metric | Limit | Tool |
|--------|-------|------|
| Function length | 60 lines | `funlen` |
| File length | 300 lines | `funlen` |
| Cyclomatic complexity | 15 | `cyclop` |

Test files (`_test.go`) are excluded from all checks.

## golangci-lint Configuration

A `.golangci.yml` at the repo root:

- **Linters enabled:** `funlen`, `cyclop`, `staticcheck`, `govet`
- **funlen settings:** `lines: 60`, `statements: 40`
- **cyclop settings:** `max-complexity: 15`
- **Exclusions:** `_test.go` files excluded from `funlen` and `cyclop`

This replaces the standalone `go vet` and `staticcheck` CI steps with a single `golangci-lint` invocation.

## CI Workflow Changes

Replace the current separate `Vet` and `Staticcheck` steps in `.github/workflows/ci.yml` with:

```yaml
- name: Lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
```

The action handles installation and caching. The final CI steps become:

1. Build (`go build ./...`)
2. Test (`go test ./... -v -race`)
3. Lint (`golangci-lint` via action)

## Refactoring Plan

### `internal/tui/app.go` (456 lines, `Update()` 167 lines)

- Extract message handler methods from `Update()`: `handleKeyMsg()`, `handleStreamChunk()`, `handleScreenMsg()`, etc. `Update()` becomes a thin type-switch dispatcher.
- Move `waitForStreamMessage()` (65 lines) to a new `stream.go` file.
- `app.go` retains: `NewApp()`, `Update()` (dispatcher), `View()`, `handleSubmit()`.

### `internal/agents/worker.go` (451 lines, `Run()` 77 lines)

- Split into `worker.go` (core struct, `Run()` loop, state management) and `worker_tools.go` (tool extraction, diff generation, file path helpers: `extractToolFilePath()`, `generateGitDiff()`, `extractFilePath()`, `buildInstructions()`).

### `internal/tui/screens/chat.go` (287 lines, `Update()` 120 lines)

- File is under the 300-line limit but `Update()` exceeds 60 lines.
- Extract handler methods: `handleKeyMsg()`, `handleStreamMsg()`, `handleToolMsg()`, etc. `Update()` becomes a type-switch dispatcher.

## Success Criteria

- `golangci-lint run` passes with zero violations
- No source file exceeds 300 lines
- No function exceeds 60 lines
- No function exceeds cyclomatic complexity of 15
- CI runs the new lint step on all pushes and PRs to main
