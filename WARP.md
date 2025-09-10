# WARP.md

This file provides guidance to Warp (warp.dev) when working with code in this repository.

Overview

- **Ghost** is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. It is designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

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
- Technical designs must maintain a strong separation of concerns to keep the project easy to maintain and extend.

Notes

- As code under ./cmd/assistant and additional packages are added, update this file to reflect the concrete commands and architecture.

---

## Quick Start (Planned MVP)

Build all:

```bash
go build ./...
```

Build binary explicitly:

```bash
go build -o ./bin/assistant ./cmd/assistant
```

Run with argument vs stdin:

```bash
go run ./cmd/assistant -model llama3.1 "Hello"
echo "Hello via stdin" | go run ./cmd/assistant -model llama3.1
```

PowerShell equivalents:

```powershell
go run ./cmd/assistant -model llama3.1 "Hello"
"Hello via stdin" | go run ./cmd/assistant -model llama3.1
```

Set environment (PowerShell vs POSIX):

```powershell
$env:OLLAMA_MODEL = "llama3.1"; go run ./cmd/assistant "Hi"
```

```bash
export OLLAMA_MODEL=llama3.1; go run ./cmd/assistant "Hi"
```

Streaming enabled by default (future `-stream=false` to disable).

---

## Environment Variables

| Variable     | Purpose               | Required                   | Default                  |
| ------------ | --------------------- | -------------------------- | ------------------------ |
| OLLAMA_MODEL | Model name for Ollama | Yes (unless flag provided) | None                     |
| OLLAMA_HOST  | Base URL for Ollama   | No                         | <http://localhost:11434> |
| LOG_FORMAT   | `json` for JSON logs  | No                         | `http://localhost:11434` |
| LOG_FORMAT   | `json` for JSON logs  | No                         | text                     |

Planned (not active yet): `GHOST_CARD` for persona selection.

---

## Exit Codes (Planned)

| Code | Meaning                |
| ---- | ---------------------- |
| 0    | Success                |
| 1    | Runtime error          |
| 2    | Invalid usage / config |
| 3    | Model unavailable      |
| 4    | Tool failure / denied  |
| 5    | Canceled / timeout     |

---

## Logging

- Uses `slog` (text by default; JSON when `LOG_FORMAT=json`).
- Model output to stdout; logs to stderr (enables piping output cleanly).

---

## Testing & Quality

Test-Driven Development (TDD) is preferred: write tests before or alongside implementation. This means writing a failing test first ('red'), then implementing code to make it pass ('green'), and finally refactoring while keeping tests passing.

Use sentinel errors (package-level variables, e.g., `var ErrModelEmpty = errors.New("model cannot be empty")`) for robust error handling and testing. Wrap sentinel errors with `%w` and use `errors.Is` for assertions in tests.

Example:

```go
// internal/llm/errors.go
package llm
import "errors"
var ErrModelUnavailable = errors.New("model unavailable")

// internal/app/app.go
if model == "" {
  return nil, fmt.Errorf("app init: %w", ErrModelEmpty)
}

// internal/llm/ollama.go
if !modelAvailable {
  return nil, fmt.Errorf("llm: %w", ErrModelUnavailable)
}
```

Run all tests:

```bash
go test ./...
```

Race & vet (optional early):

```bash
go vet ./...
go test -race ./...
```

Each boundary package should expose an interface + mock/fake for consumption in tests of downstream packages.

Planned: `golangci-lint` integration.

---

## Troubleshooting

| Symptom            | Likely Cause                     | Resolution                            |
| ------------------ | -------------------------------- | ------------------------------------- |
| Connection refused | Ollama not running               | Start Ollama server/app               |
| Model not found    | Model not pulled                 | `ollama pull llama3.1`                |
| Empty output       | Wrong model / host               | Verify `OLLAMA_MODEL` & `OLLAMA_HOST` |
| Exit code 3        | Host unreachable / model missing | Confirm host & model list             |

Check environment quickly:

```bash
echo $OLLAMA_MODEL
curl -s http://localhost:11434/api/tags | jq '.models[]?.name' # optional
```

PowerShell:

```powershell
echo $env:OLLAMA_MODEL
```

---

## Planned Structure (Subject to Change)

```text
cmd/assistant/main.go        # CLI entrypoint (MVP)
internal/llm                 # Ollama client + streaming interfaces
internal/tools               # Tool interface + (MVP) web search tool
internal/memory              # Added in later phases (session, long-term)
internal/config              # Flag/env resolution
cards/                       # Persona Markdown cards
```

---

## Web Search Tool (Planned)

- Only tool in MVP besides core chat.
- Failure or denial may map to exit code 4.
  -- Provider undecided (intentionally deferred).

---

## Cards (Personas)

- Markdown files (optional YAML front matter) under `cards/`.
- Future: `-card` flag or `GHOST_CARD` env.
- Fallback: neutral embedded system prompt if none provided.

---

## Concurrency & Cancellation

- All long-running operations accept `context.Context`.
- User interrupts (Ctrl+C) should cancel stream (exit code 5 planned).

---

## Deferred / Out of Scope (MVP)

| Item                  | Reason                              |
| --------------------- | ----------------------------------- |
| TUI                   | Post-MVP milestone                  |
| Vector / embeddings   | Introduce with memory Phase 2       |
| Multi-LLM backends    | Ollama-only simplifies early design |
| Config file           | Flags/env sufficient                |
| Shell/code exec tools | Reduce early security surface       |

---

## Security (Early Posture)

- No file system mutations beyond standard operation.
- API keys (future) remain in env only (never logged).

---

## Future Markers

- Add TUI (Bubble Tea) streaming interface.
- Introduce memory persistence path (likely user config directory).
- Add `golangci-lint` config.
