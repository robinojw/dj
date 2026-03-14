# Complexity CI Checks Design

## Overview

Add code complexity enforcement to CI using `golangci-lint` with `funlen` and `cyclop` linters. Enforce file length via a simple CI script. Refactor existing code that exceeds thresholds. Update `CLAUDE.md` to reflect the new lint commands.

## Thresholds

| Metric | Limit | Enforced by |
|--------|-------|-------------|
| Function length | 60 lines | `funlen` linter |
| File length | 300 lines (source only) | Shell script in CI |
| Cyclomatic complexity | 15 | `cyclop` linter |

Test files (`_test.go`) are excluded from all checks.

## golangci-lint Configuration

Uses the v2 configuration format. `golangci-lint` v2 requires action `@v7` or later. Exact `.golangci.yml`:

```yaml
version: "2"

linters:
  default: none
  enable:
    - govet
    - staticcheck
    - funlen
    - cyclop

  settings:
    funlen:
      lines: 60
      statements: -1
    cyclop:
      max-complexity: 15
    staticcheck:
      checks:
        - "all"
        - "-QF*"

  exclusions:
    rules:
      - path: '(.+)_test\.go'
        linters:
          - funlen
          - cyclop
```

Using `default: none` with explicit enables avoids inheriting linters like `errcheck` that have pre-existing violations outside this project's scope. `govet` and `staticcheck` are explicitly enabled to maintain existing CI coverage. The `statements: -1` disables funlen's statement count check — only line count is enforced. The `staticcheck` QF (quickfix) checks are disabled because the current standalone staticcheck action does not enable them — this preserves existing behavior.

Note: Run `golangci-lint run` locally after configuration to confirm exact violations, as `cyclop` may count complexity slightly differently from `gocyclo`.

## File Length Check

`funlen` does not check file length. A simple shell script step in CI enforces the 300-line limit on non-test source files:

```yaml
- name: Check file length
  run: |
    violations=""
    for f in $(find . -name '*.go' ! -name '*_test.go' -not -path './vendor/*'); do
      lines=$(wc -l < "$f")
      if [ "$lines" -gt 300 ]; then
        violations="${violations}${f}: ${lines} lines (max 300)\n"
      fi
    done
    if [ -n "$violations" ]; then
      printf "File length violations:\n%b" "$violations"
      exit 1
    fi
```

## CI Workflow Changes

Replace the current separate `Vet` and `Staticcheck` steps in `.github/workflows/ci.yml` with:

```yaml
- name: Lint
  uses: golangci/golangci-lint-action@v7
  with:
    version: v2.1.6

- name: Check file length
  run: |
    violations=""
    for f in $(find . -name '*.go' ! -name '*_test.go' -not -path './vendor/*'); do
      lines=$(wc -l < "$f")
      if [ "$lines" -gt 300 ]; then
        violations="${violations}${f}: ${lines} lines (max 300)\n"
      fi
    done
    if [ -n "$violations" ]; then
      printf "File length violations:\n%b" "$violations"
      exit 1
    fi
```

The action handles installation and caching. The final CI steps become:

1. Build (`go build ./...`)
2. Test (`go test ./... -v -race`)
3. Lint (`golangci-lint` via action)
4. Check file length (shell script)

## Documentation Updates

Update `CLAUDE.md` lint commands from:
```
go vet ./...
staticcheck ./...
```
to:
```
golangci-lint run
```

## Refactoring Plan

### `internal/tui/app.go` (456 lines, `Update()` ~166 lines, `waitForStreamMessage()` ~65 lines)

- Extract message handler methods from `Update()`: `handleKeyMsg()`, `handleStreamChunk()`, `handleScreenMsg()`, etc. `Update()` becomes a thin type-switch dispatcher.
- Move `waitForStreamMessage()` (~65 lines, cyclomatic complexity ~18) to a new `stream.go` file and decompose its nested select logic into smaller helpers.
- `app.go` retains: `NewApp()`, `Update()` (dispatcher), `View()`, `handleSubmit()`.

### `internal/agents/worker.go` (451 lines, `Run()` ~73 lines, `streamResponse()` ~69 lines)

- Split into `worker.go` (core struct, `Run()` loop, state management) and `worker_tools.go` (tool extraction, diff generation, file path helpers: `extractToolFilePath()`, `generateGitDiff()`, `extractFilePath()`, `buildInstructions()`).
- Decompose `streamResponse()` — extract chunk-type handling into helper methods to bring it under 60 lines.

### `internal/tui/screens/chat.go` (`Update()` ~120 lines, complexity ~20)

- File is under the 300-line limit but `Update()` exceeds both the 60-line funlen limit and the cyclomatic complexity threshold of 15.
- Extract handler methods: `handleKeyMsg()`, `handleStreamMsg()`, `handleToolMsg()`, etc. `Update()` becomes a type-switch dispatcher.

### `cmd/harness/main.go` (`main()` ~102 lines)

- Extract initialization logic into helper functions: `initConfig()`, `initMCP()`, `initApp()`, etc. `main()` becomes a sequence of high-level calls.

### `internal/tools/builtin.go` (`registerBuiltinSchemas()` ~91 lines)

- Split schema registrations into grouped helpers by tool category (file tools, search tools, agent tools, etc.), or use a table-driven approach with a loop.

### `internal/tui/components/debug_overlay.go` (`View()` ~76 lines)

- Extract section renderers: `renderToolCalls()`, `renderTokenUsage()`, `renderMessages()`, etc.

### `internal/agents/dag.go` (`buildDAG()` ~67 lines, complexity ~16)

- Triggers both `funlen` (length) and `cyclop` (complexity). Extract subtask parsing and dependency resolution into separate functions.

### `internal/agents/orchestrator.go` (`Dispatch()` ~65 lines)

- Extract task analysis and worker spawning into helper methods.

### `internal/tui/screens/cheat_sheet.go` (`View()` ~62 lines)

- Extract shortcut table and mode table rendering into helper functions.

## Success Criteria

- `golangci-lint run` passes with zero violations on source files
- No source file exceeds 300 lines
- No function exceeds 60 lines
- No function exceeds cyclomatic complexity of 15
- CI runs lint and file-length steps on all pushes and PRs to main
- `CLAUDE.md` reflects the updated lint commands
