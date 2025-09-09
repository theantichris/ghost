# WARP.md

This file provides guidance to Warp (warp.dev) when working with code in this repository.

Overview

- assistant-go is a local conversational AI assistant and orchestrator built in Go, powered by a locally running Ollama instance.

Prerequisites

- Go >= 1.24
- Ollama installed and running locally
- At least one model pulled (e.g., `ollama pull llama3.1`)

Core commands (when implemented)

- Build CLI binary
  - go build -o ./bin/assistant ./cmd/assistant
- Build all packages (used by CI)
  - go build ./...
- Run without building
  - go run ./cmd/assistant -model MODEL "PROMPT"
  - echo "Hello" | go run ./cmd/assistant -model MODEL
- Tests
  - Run all tests: go test ./...
  - Run a single test by name (example): go test ./... -run '^TestName$' -v

Configuration (flags and env; planned)

- Flags
  - -model string (required unless env OLLAMA_MODEL is set)
  - -stream bool (default true)
- Environment variables
  - OLLAMA_MODEL
  - OLLAMA_HOST

Big-picture architecture (high level)

- See [SPEC.md](SPEC.md) for architecture and [ROADMAP.md](ROADMAP.md) for milestones.
- Initial entry point will be a CLI at ./cmd/assistant. Input via CLI args or stdin, a chat call to Ollama’s /api/chat with streaming output. A TUI (Bubble Tea) is planned for an interactive chat view.

CI

- GitHub Actions workflow (.github/workflows/ci.yml) runs on push/PR to main:
  - Sets up Go 1.24.x
  - go build ./...
  - go test ./...

Tooling/Rules

- GitHub Copilot guidance is at [.github/copilot-instructions.md](.github/copilot-instructions.md). Follow [SPEC.md](SPEC.md) and [ROADMAP.md](ROADMAP.md) when proposing or generating code.
- Local-first, privacy-first. Any networked tools must be explicitly opt-in and clearly documented.

Notes

- As code under ./cmd/assistant and additional packages are added, update this file to reflect the concrete commands and architecture.
