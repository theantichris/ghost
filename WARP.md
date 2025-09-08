# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Overview

- assistant-go is a local conversational AI assistant and orchestrator built in Go, powered by a locally running Ollama instance.
- Current status: MVP CLI named "assistant" sends a single user message to Ollama’s /api/chat endpoint and prints the assistant response. Streaming is enabled by default.

Prerequisites

- Go >= 1.24
- Ollama installed and running locally (default <http://localhost:11434>)
- At least one model pulled (e.g., `ollama pull llama3.1`)

Core commands

- Build CLI binary
  - go build -o ./bin/assistant ./cmd/assistant
- Build all packages (used by CI)
  - go build ./...
- Run without building
  - go run ./cmd/assistant -model MODEL "PROMPT"
  - echo "Hello" | go run ./cmd/assistant -model MODEL
  - Using environment variables
    - export OLLAMA_MODEL=llama3.1; export OLLAMA_HOST=<http://localhost:11434>; go run ./cmd/assistant "Hello from env"
- Tests
  - Run all tests: go test ./...
  - Run a single test by name (example): go test ./... -run '^TestName$' -v

Configuration (flags and env)

- Flags
  - -model string (required unless env OLLAMA_MODEL is set)
  - -host string (default from env OLLAMA_HOST or <http://localhost:11434>)
  - -timeout dur (default 2m)
  - -stream bool (default true)
- Environment variables
  - OLLAMA_MODEL
  - OLLAMA_HOST

Big-picture architecture

- Entry point: CLI at ./cmd/assistant (as referenced in README). Accepts prompt via CLI args or stdin, performs a single-turn chat call to Ollama’s /api/chat, and streams tokens to stdout. No additional internal packages or services are described in the repository at this time.

CI

- GitHub Actions workflow (.github/workflows/ci.yml) runs on push/PR to main:
  - Sets up Go 1.24.x
  - go build ./...
  - go test ./...

Tooling/Rules

- No CLAUDE.md, Cursor rules (.cursor/ or .cursorrules), or Copilot instructions (.github/copilot-instructions.md) are present.
- .vscode/settings.json contains only a cSpell word for "Ollama".

Notes

- If/when code under ./cmd/assistant and additional packages are added, update the Big-picture architecture and commands accordingly.
