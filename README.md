# DJ

An AI coding harness TUI powered by OpenAI Codex.

DJ orchestrates AI agents in your terminal — plan, build, refactor, and debug code through an interactive interface with built-in skills, MCP server support, and LSP integration.

## Install

### Homebrew

```bash
brew install robinojw/dj
```

### From source

```bash
go install github.com/robinojw/dj/cmd/harness@latest
```

## Setup

DJ requires an OpenAI API key:

```bash
export OPENAI_API_KEY="your-key-here"
```

## Usage

```bash
dj
```

This launches the TUI. From there you can chat with the AI agent, use built-in skills, and orchestrate multi-step coding tasks.

```bash
dj --version
```

## Configuration

DJ reads configuration from `harness.toml` in your project root, with user-level overrides from `~/.config/codex-harness/config.toml`.

## License

[MIT](LICENSE)
