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

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/ghost.svg)](https://pkg.go.dev/github.com/theantichris/ghost) [![Build Status](https://github.com/theantichris/ghost/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/ghost/actions) [![Build Status](https://github.com/theantichris/ghost/actions/workflows/markdown.yml/badge.svg)](https://github.com/theantichris/ghost/actions) [![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost) ![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

**Ghost** is a local, general-purpose AI assistant CLI tool built in Go and powered by Ollama. It provides command-line access to AI capabilities for quick queries, code analysis, and task automation, running entirely on your own machine.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_, _Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI companion into a terminal-first experience.

## Documents

- Specification: [SPEC.md](SPEC.md)
- Agents: [AGENTS.md](AGENTS.md)

## Requirements

- Go >= 1.24
- Ollama installed and running locally
- At least one model pulled (e.g., `ollama pull llama3.1`)

## Usage

```bash
# Ask a single question
ghost ask "What is the capital of France?"

# Pipe input to Ghost for analysis
cat code.go | ghost ask "Explain this code"

# Combine piped input with additional context
cat error.log | ghost ask "What's causing this error?"

# Run directly with go
go run main.go ask "Your question here"
```

Ghost processes your query through the configured LLM model and returns the response directly to stdout. Responses stream in real-time as tokens are generated.

### Commands

#### `ask` - Ask Ghost a question

Send a query to the LLM and get a response. Supports both direct queries and piped input.

**Command-specific flags:**

- `--no-newline, -n` — Don't add newline after response (useful for scripts)
- `--timeout` — HTTP timeout for LLM requests (default: 2 minutes)

### Global Flags

- `--model` — Override the default LLM model
- `--debug` — Enable verbose diagnostics and DEBUG level logging
- `--config` — Specify config file location (default: `$HOME/.ghost.yaml`)

### Environment Variables

- `OLLAMA_BASE_URL` — base URL of your local Ollama instance (e.g., `http://127.0.0.1:11434`).
- `DEFAULT_MODEL` — default model name to use when `-model` is not provided.

### Configuration

Ghost checks for configuration in the following order of precedence:

1. Command-line flags
2. Environment variables
3. Config file (`.ghost.yaml`)
4. Default values

Configuration options:

- Model selection via `--model` flag, `DEFAULT_MODEL` env var, or config file
- Ollama base URL via `OLLAMA_BASE_URL` env var or config file

## Features

- **CLI Command Interface**: Clean command-line interface using Cobra framework
- **Streaming Responses**: Real-time token-by-token output as the model generates responses
- **Pipe Support**: Process files, logs, or command output by piping to Ghost
- **Think Block Filtering**: Automatically filters out `<think>` blocks from model responses
- **Flexible Configuration**: Support for environment variables, config files, and command-line flags
- **Structured Logging**: Clean, component-based logging with adjustable verbosity
