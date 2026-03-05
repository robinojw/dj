# Release Setup Design

## Goal

Set up a GoReleaser-based release pipeline that cross-compiles `dj`, publishes GitHub releases with changelogs, and auto-updates a Homebrew tap.

## Components

### 1. Version wiring (`cmd/harness/main.go`)

Add `version`, `commit`, `buildDate` package-level vars populated via ldflags at build time. Add `--version` flag handling before TUI launch.

### 2. `.goreleaser.yml`

- Build `./cmd/harness/main.go` as binary `dj`
- Targets: linux/darwin x amd64/arm64, windows/amd64
- Archives include README.md, LICENSE, themes/, skills/
- Changelog grouped by conventional commit prefix
- Homebrew tap auto-pushes to `robinojw/homebrew-dj`

### 3. `.github/workflows/release.yml`

Triggers on `v*` tags. Runs GoReleaser with `GITHUB_TOKEN` (release creation) and `GH_PAT` (tap repo push).

### 4. `robinojw/homebrew-dj` repo

New GitHub repo with `Formula/` directory. GoReleaser manages `Formula/dj.rb` automatically.

### 5. Supporting files

- `LICENSE` — MIT
- `README.md` — project description, install instructions, basic usage

## Out of scope

- Curl install script
- Windows ARM64 target

## Manual steps after implementation

1. Create a fine-grained GitHub PAT scoped to `homebrew-dj` repo with Contents write permission
2. Add `GH_PAT` as a repository secret in `robinojw/dj`
