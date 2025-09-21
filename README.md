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

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/ghost.svg)](https://pkg.go.dev/github.com/theantichris/ghost) [![Build Status](https://github.com/theantichris/ghost/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/ghost/actions) [![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost) ![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

**Ghost** is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. It is designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

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
go run ./cmd/ghost
```

Ghost seeds every session with its system prompt, greets you on startup, and maintains in-memory chat history for context. Responses stream in real-time as tokens are generated. Type messages directly into the terminal; send `/bye` to end the session.

### Flags

- `-model` — overrides `DEFAULT_MODEL`; if omitted, Ghost falls back to the environment variable.
- `-debug` — enables verbose diagnostics. When enabled, Ghost logs at DEBUG level and dumps the chat history to stdout after the session for quick inspection.

### Environment Variables

- `OLLAMA_BASE_URL` — base URL of your local Ollama instance (e.g., `http://127.0.0.1:11434`).
- `DEFAULT_MODEL` — default model name to use when `-model` is not provided.

### Recoverable Errors

Temporary issues (network hiccups, Ollama downtime, malformed responses) surface as system messages inside the chat instead of exiting immediately, so you can retry once the condition clears.

## Features

- **Streaming Responses**: Real-time token-by-token output as the model generates responses
- **Think Block Filtering**: Automatically filters out `<think>` blocks from thinking models while preserving them in chat history
- **Graceful Error Recovery**: Network issues and API errors display as system messages without ending the session
