# Agents.md Discovery

## Summary

Add support for discovering and loading `AGENTS.md` files from anywhere in the repository, not just the project root. All discovered files are concatenated into the context sent with every API request.

## Requirements

- At app startup, find all `agents.md` files in the repository (case-insensitive)
- Discovery uses `find . -iname "agents.md" -not -path "./.git/*"` via shell execution
- Files are ordered depth-first from root, then alphabetically within the same depth
- All discovered files are always loaded and injected as context
- Falls back to `["AGENTS.md"]` if the shell command fails

## Design

### Discovery (`main.go`)

At startup, before creating the memory manager:

1. Run `find . -iname "agents.md" -not -path "./.git/*"` via `exec.Command` from the working directory
2. Parse stdout lines into a `[]string` of relative paths, normalizing any `./` prefix
3. Sort by depth (count of `/` separators), then alphabetically within same depth
4. Pass the sorted slice to `memory.NewManager(paths, userPath)`

If `find` fails, fall back to `[]string{"AGENTS.md"}` to preserve existing behavior.

### Memory Manager (`internal/memory/manager.go`)

Change the `Manager` struct to hold multiple project paths:

```go
type Manager struct {
    projectPaths []string   // was: projectPath string
    userPath     string
}
```

`NewManager(projectPaths []string, userPath string)` replaces the current single-path constructor. `DefaultManager()` stays as a fallback using `[]string{"AGENTS.md"}`.

`LoadContext()` iterates the sorted paths and builds:

```
<project_memory>
--- AGENTS.md (./AGENTS.md) ---
<root content>

--- AGENTS.md (./internal/tools/AGENTS.md) ---
<subdirectory content>
</project_memory>

<user_memory>
...
</user_memory>
```

Files that don't exist or are empty are silently skipped. `ProjectPath()` becomes `ProjectPaths() []string`.

### Chat Screen Wiring (`internal/tui/app.go`)

The `App` struct currently has no `memory` field. Changes needed:

1. Add `memory *memory.Manager` field to `App`
2. Update `NewApp` to accept a `*memory.Manager` parameter
3. In `handleSubmit`, change `Instructions` from `modeCfg.SystemPrompt` to `modeCfg.SystemPrompt + "\n\n" + a.memory.LoadContext()`

### Worker Pipeline Wiring

The `Worker` already has a `memory` field and `buildInstructions()` calls `w.memory.LoadContext()`. The `Orchestrator` has an exported `Memory` field that must be set after construction in `main.go` to pass the manager through to workers.

No configuration file changes needed. Discovery is automatic.

## Files Changed

| File | Change |
|------|--------|
| `cmd/harness/main.go` | Add discovery logic, update `NewManager` call, pass manager to `NewApp` and orchestrator |
| `internal/memory/manager.go` | `projectPath string` -> `projectPaths []string`, update `LoadContext()`, update constructors |
| `internal/memory/manager_test.go` | Update tests for new multi-path constructor signature |
| `internal/tui/app.go` | Add `memory` field to `App`, update `NewApp` signature, use memory context in `handleSubmit` |

## Out of Scope

- Configurable filename patterns (hardcoded to `agents.md`)
- Mid-session refresh (load once at startup only)
- Caching or file watching
- Symlink following (avoids cycle risk)
- Size limits on discovered files
