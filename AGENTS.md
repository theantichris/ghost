# AGENTS.md

Purpose: Help automated coding agents make correct, minimal, forward-compatible changes.

Always align generated code and docs with:

- [SPEC.md](SPEC.md)
- [ROADMAP.md](ROADMAP.md)

## Project Overview

A local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. Designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_, _Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI companion into a terminal-first experience.

## Core Principles

- **Cross-platform**: Support Windows, macOS, Linux with portable paths/shells/filesystems
- **Small, composable components**: Clear interfaces, predictable flags and output
- **Strong separation of concerns**: Keep the project easy to maintain and extend
- **Test-driven development**: Write failing tests before implementing code

## Architecture

### Package Structure

```text
cmd/ghost/           CLI entrypoint
internal/llm/        Ollama client, streaming
internal/tui/        Bubble Tea chat view (planned)
internal/tools/      Tool framework and built-ins (planned)
internal/memory/     Session/history, RAG integration (planned)
internal/config/     Flag/env/config precedence (planned)
pkg/                 Only if stable external API needed
```

### Key Patterns

- Each boundary returns wrapped errors upward; only CLI maps to exit codes
- Internal packages include interface + mock for testing
- Future features add new interfaces rather than mutating existing ones
- Keep HTTP client reuse (no per-call instantiation)
- Context as first param (after receiver) for cancelable operations; never store in structs

## Technology Stack

**Core**: Go 1.24 + standard library (`net/http`, `context`, `log/slog`)

**Dependencies**:

- `github.com/joho/godotenv` (optional .env loading)
- `github.com/MatusOllah/slogcolor` (CLI log coloring only)

**External runtime**: Local Ollama server (HTTP API at `OLLAMA_BASE_URL`)

## Code Standards

### Style & Quality

- Use `gofmt` (CI enforces)
- Standard imports grouping: stdlib / third-party / internal
- Keep functions small; avoid premature abstraction
- Sentinel errors (`ErrURLEmpty`, etc.) + `%w` wrapping at boundaries
- Prefer explicit error handling over cleverness
- Avoid new dependencies unless essential

### Error Handling & Exit Codes

| Code | Meaning                        |
| ---- | ------------------------------ |
| 0    | Success                        |
| 1    | Runtime error                  |
| 2    | Invalid usage / configuration  |
| 3    | Model unavailable / host issue |
| 4    | Tool failure / denied          |
| 5    | Canceled / timeout             |

Example error pattern:

```go
// Define sentinel errors
var ErrModelUnavailable = errors.New("model unavailable")

// Wrap at boundaries
if !modelAvailable {
    return nil, fmt.Errorf("llm: %w", ErrModelUnavailable)
}

// Test with errors.Is
if errors.Is(err, llm.ErrModelUnavailable) {
    // handle
}
```

### Logging & Output

- Use `log/slog` exclusively with structured fields
- Always set `component` field: `main`, `app`, `ollama client`
- **stdout reserved strictly for model/user chat output**
- **Logs go to stderr** to avoid accidental leakage in pipelines
- Avoid logging secrets, API keys, raw env values

### Concurrency & Context

- Every blocking method takes `context.Context` as first parameter
- Document goroutines: purpose, ownership, cancellation path
- Avoid unbounded buffering; prefer streaming with backpressure
- Ensure all goroutines terminate on context cancellation

### Security

- No credential logging; redact or omit sensitive data
- No telemetry, analytics, or remote calls
- Maintain separation: logs (stderr), model output (stdout)

## Testing

### Guidelines

- **Test-driven development**: Write failing tests first ('red'), implement ('green'), refactor
- Use subtests + `t.Parallel()` where independent
- Prefer mocks/fakes over real HTTP; use `httptest` only for HTTP behavior testing
- Assert sentinel errors with `errors.Is`; substring checks only for HTTP/JSON messages
- Keep tests deterministic; no reliance on environment or live Ollama
- Add tests for new error paths when introducing validation

### Commands

```bash
go test ./...              # Full test suite
go vet ./...              # Static analysis
go test -race ./...       # Race detection
```

## Build & Deployment

### Local Development

```bash
go build ./...                                    # Build all packages
go build -o ./bin/ghost ./cmd/ghost              # Build binary
go run ./cmd/ghost -model qwen3:8b               # Run (ensure Ollama running)
```

### Environment

- Provide `.env` with required variables or pass `-model` flag
- No containers or release automation defined yet
- Local usage only currently

## Development Guidelines

### Commit Style

- Imperative tense: "Add streaming client" / "Fix card parsing error"
- Keep changes minimal, aligned with current surface
- Add/update tests for new code paths
- Avoid stylistic churn; don't reformat unrelated files

### Common Gotchas

- Forgetting env vars produces early exit (handled in `main` with exit codes 2 or 3)
- Chat loop blocking in tests: inject controlled reader (pass `input` to `Run`)
- Don't move exit code logic into internal packages

## External Integrations

- **Ollama HTTP API** (`/api/chat` endpoint)
  - Reference: <https://github.com/ollama/ollama> (context only; don't fetch during tests)
