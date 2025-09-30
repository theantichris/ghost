# AGENTS.md

Purpose: Help automated coding agents make correct, minimal, forward-compatible changes.

Always align generated code and docs with:

- [SPEC.md](SPEC.md)

## Project Overview

A local, general-purpose AI assistant CLI tool built in Go and powered by
Ollama. Provides command-line access to AI capabilities for quick queries,
code analysis, and task automation, running entirely on your own machine.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI
companion into a terminal-first experience.

This repo also serves as the maintainer's personal sandbox for relearning
modern Go after time away from coding. Answer questions like a tutorial aimed
and teaching the user. Provide explanations and incremental, well-referenced
changes to support that learning goal.

## Core Principles

- **Cross-platform**: Support Windows, macOS, Linux with portable paths/shells/filesystems
- **Small, composable components**: Clear interfaces, predictable flags and output
- **Strong separation of concerns**: Keep the project easy to maintain and extend
- **Test-driven development**: Write failing tests before implementing code

## Architecture

### Package Structure

```text
cmd/                 Root command and subcommands
cmd/root.go         Root command setup, config initialization
cmd/ask.go          Ask command implementation
internal/llm/        Ollama client with streaming implementation
  - client.go       LLMClient interface definition
  - ollama.go       OllamaClient implementation
  - errors.go       Domain-specific error definitions
internal/tui/        Bubble Tea chat view (planned)
internal/tools/      Tool framework and built-ins (planned)
internal/memory/     Session/history, RAG integration (planned)
pkg/                 Only if stable external API needed
```

### Key Patterns

- Each boundary returns wrapped errors upward; only CLI maps to exit codes
- Internal packages include interface + mock for testing (e.g., `LLMClient`
  interface with `MockLLMClient`)
- Command layer handles configuration validation, internal packages handle
  domain validation
- Dual validation pattern: config layer validates presence, domain layer
  validates correctness
- Component-based logging with structured fields (always include `component` field)
- Future features add new interfaces rather than mutating existing ones
- Constructors that need optional behavior accept a config/options struct
  (e.g., output writers, debug toggles) instead of long positional argument
  lists
- Keep HTTP client reuse (no per-call instantiation)
- Context as first param (after receiver) for cancelable operations; never
  store in structs
- Streaming responses use callback functions (`onToken func(string)`) for
  real-time output
- Think block filtering strips `<think>...</think>` tags from model output

## Technology Stack

**Core**: Go 1.24 + standard library (`net/http`, `context`, `log/slog`)

**Dependencies**:

- `github.com/charmbracelet/fang` (CLI enhancement framework)
- `github.com/spf13/cobra` (CLI command framework)
- `github.com/spf13/viper` (configuration management)
- `github.com/joho/godotenv` (optional .env loading)
- `github.com/MatusOllah/slogcolor` (CLI log coloring only)

**External runtime**: Local Ollama server (HTTP API at `OLLAMA_BASE_URL`)

## Code Standards

### Style & Quality

- Use `gofmt` (CI enforces)
- Standard imports grouping: stdlib / third-party / internal
- Keep functions small; avoid premature abstraction
- Sentinel errors (`ErrURLEmpty`, etc.) + `%w` wrapping at boundaries
- Hoist user-facing strings (labels, commands, prompts) into shared constants
  so code and tests stay aligned.
- Avoid magic strings; promote shared literals (messages, prompts, flags) to
  constants for reuse across code, tests, and docs.
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

Example error patterns:

```go
// Sentinel error definition
var ErrClientResponse = errors.New("failed to get response from Ollama API")
var ErrURLEmpty = errors.New("OLLAMA_BASE_URL not configured")

// Logging with structured fields
Logger.Error("couldn't initialize LLM client",
    "error", ErrModelEmpty,
    "component", "cmd.askCmd.initializeLLMClient")

// Wrapping at boundary
if err != nil {
    return "", fmt.Errorf("%w: %w", ErrResponse, err)
}

// Wrapping with extra context
if statusCode/100 != 2 {
    return "", fmt.Errorf("%w: status=%d %s body=%q", ErrResponse,
        statusCode, http.StatusText(statusCode), string(responseBody))
}

// User-friendly error messages in command layer
return nil, fmt.Errorf(
    "%w, set it via OLLAMA_BASE_URL environment variable or config file",
    ErrURLEmpty)

// Checking in consumer
if errors.Is(err, llm.ErrClientResponse) {
    // handle error
}
```

### Logging & Output

- Use `log/slog` exclusively with structured fields
- Always set `component` field with specific context: `cmd.askCmd.runAsk`,
  `llm.OllamaClient.Chat`, etc.
- **stdout reserved strictly for command output (model responses)**
- **Logs go to stderr** to avoid accidental leakage in pipelines
- Error logging happens at the appropriate layer:
  - Command layer: user-facing errors with remediation hints
  - Internal packages: domain-specific errors with technical context
- Avoid duplicate logging of the same error across layers
- Personal debug dumps (e.g., spew) may write to stdout when guarded by an
  explicit developer-only flag
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

- **Test-driven development**: Write failing tests first ('red'), implement
  ('green'), refactor
- Use subtests + `t.Parallel()` where independent
- Prefer mocks/fakes over real HTTP; use `httptest` only for HTTP behavior
  testing
- Assert sentinel errors with `errors.Is`; substring checks only for HTTP/JSON messages
- Keep tests deterministic; no reliance on environment or live Ollama
- Add tests for new error paths when introducing validation
- For dependency injection in tests:
  - Extract testable functions that accept interfaces (e.g., `runAskWithClient`)
  - Accept limited test coverage on simple glue code (config/initialization)
  - Focus testing effort on complex business logic

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
go build -o ./bin/ghost main.go                  # Build binary
go run main.go ask "Your question"               # Run ask command
go run main.go ask --model qwen3:8b "Question"  # Run with specific model
cat file.go | go run main.go ask "Explain this" # Pipe input to Ghost
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

- Forgetting env vars produces descriptive error messages with configuration hints
- Duplicate validation between command and internal layers is intentional
  (separation of concerns)
- Don't move exit code logic into internal packages
- When testing commands with hard dependencies, extract testable functions
  that accept interfaces
- Component names in logging should be specific to the function/method for
  better debugging

## External Integrations

- **Ollama HTTP API** (`/api/chat` endpoint)
  - Repo: <https://github.com/ollama/ollama> (context only; don't fetch during tests)
  - API Reference: <https://github.com/ollama/ollama/blob/main/docs/api.md>
    (context only; don't fetch during tests)
