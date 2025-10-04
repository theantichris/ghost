# ghost

```text
   ▄████  ██░ ██  ▒█████   ██████  ████████
  ██▒ ▀█▒▓██░ ██▒▒██▒  ██▒ ██    ▒    ██
 ▒██░▄▄▄░▒██▀▀██░▒██░  ██▒ ▓██▄       ██
 ░▓█  ██▓░▓█ ░██ ▒██   ██░  ▒   ██▒   ██
 ░▒▓███▀▒░▓█▒░██▓░ ████▓▒░▒██████▒▒   ██
  ░▒   ▒   ▒ ░░▒░▒░ ▒░▒░▒░ ▒ ▒▓▒ ▒ ░   ░
   ░   ░   ▒ ░▒░ ░  ░ ▒ ▒░ ░ ░▒  ░ ░
 ░ ░   ░   ░  ░░ ░░ ░ ░ ▒  ░  ░  ░
       ░   ░  ░  ░    ░ ░        ░
```

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/ghost.svg)](https://pkg.go.dev/github.com/theantichris/ghost)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/markdown.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost)
![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

**Ghost** is a local, general-purpose AI assistant CLI tool built in Go and
powered by Ollama. It provides command-line access to AI capabilities for
quick queries, code analysis, and task automation, running entirely on your
own machine.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI
companion into a terminal-first experience.

## Documents

- Specification: [SPEC.md](SPEC.md)
- Agents: [AGENTS.md](AGENTS.md)

## Requirements

- Go >= 1.24
- Ollama installed and running locally
- At least one model pulled (e.g., `ollama pull llama3.1`)

## Installation

### From Source

```bash
git clone https://github.com/theantichris/ghost.git
cd ghost
go build -v
```

### From Release

Download pre-built binaries for Linux, macOS, or Windows from the [releases page](https://github.com/theantichris/ghost/releases).

## Usage

```bash
# Ask a single question
ghost ask "What is the capital of France?"

# Pipe input to Ghost for analysis
cat code.go | ghost ask "Explain this code"

# Combine piped input with additional context
cat error.log | ghost ask "What's causing this error?"

# Start an interactive chat session
ghost chat

# Run directly with go
go run main.go ask "Your question here"
```

Ghost processes your query through the configured LLM model and returns the
response directly to stdout.

### Commands

#### `ask` - Ask Ghost a question

Send a query to the LLM and get a response. Supports both direct queries and
piped input.

#### `chat` - Start an interactive chat session

Enter an interactive multi-turn conversation with Ghost. Chat maintains
in-memory conversation history throughout the session.

```bash
ghost chat
```

**Chat controls:**

- Type your messages and press Enter to send
- `/bye` or `/exit` - End the chat with a goodbye message
- `Ctrl+D` (EOF) - Exit immediately
- Empty input is ignored (just press Enter again)

### Global Flags

- `--ollama` — Override the Ollama API base URL.
- `--model` — Override the default LLM model
- `--config` — Specify config file location (default: `$HOME/.config/ghost/config.toml`)

### Environment Variables

- `OLLAMA_BASE_URL` — base URL of your local Ollama instance (e.g., `http://127.0.0.1:11434`).
- `DEFAULT_MODEL` — default model name to use when `-model` is not provided.

### Configuration

Ghost checks for configuration in the following order of precedence:

1. Command-line flags
2. Environment variables
3. Config file (`.ghost.toml`)
4. Default values

#### Config File Setup

Create a `config.toml` file in `~/.config/ghost/` to set default configuration:

```toml
# ~/.config/ghost/config.toml
ollama = "http://localhost:11434"
model = "llama3.1"
```

Configuration options:

- `ollama` — Ollama API base URL (default: `http://localhost:11434`)
- `model` — LLM model to use (default: `llama3.1`)

## Features

- **CLI Command Interface**: Enhanced command-line interface using Fang
  (built on Cobra framework) with styled help pages and improved user
  experience
- **Interactive Chat Mode**: Multi-turn conversations with in-memory history
  and graceful exit commands
- **Pipe Support**: Process files, logs, or command output by piping to Ghost
- **Streaming Output**: Real-time token streaming for responsive user experience
- **Think Block Filtering**: Automatically filters out `<think>` blocks from
  model responses in both ask and chat modes
- **Flexible Configuration**: Support for environment variables, config
  files, and command-line flags
- **File Logging**: All operations logged to `~/.config/ghost/ghost.log` for debugging
  and troubleshooting (sensitive data never logged)
- **Error Handling**: Comprehensive error messages with clear guidance on
  configuration issues
