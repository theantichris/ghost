# Spec

**Ghost** is a local, general-purpose AI assistant and orchestrator built in
Go and powered by Ollama. It is designed for research, chat, and task
automation, running entirely on your own machine with hybrid connectivity.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always on AI
companion into a terminal-first experience.

Capabilities should include research, web searching, helping with code,
generating images, executing tasks, setting up reminders, and chatting.

## Technical Architecture

### Core Principles

- **Local Execution**: Core AI processing runs locally via Ollama
- **Internet-Enabled**: External services (web search, APIs) allowed for
  enhanced capabilities
- **Modular Design**: Clean separation between engine, memory, tools, and interfaces
- **Strong Separation of Concerns**: Technical designs must enforce clear
  boundaries between components and responsibilities to keep the project easy
  to maintain and extend.

### System Components

#### 1. Core Engine

- **LLM Client**: Ollama API integration with tool support
- **Conversation Manager**: Handle chat flow, context windows, streaming
  - Seeds CLI sessions with the Ghost system prompt, captures the initial
    greeting before user input, maintains in-memory turn history, and exits
    on the `/bye` command.
  - Streaming responses implemented with real-time token output and think
    block filtering for thinking models.
- **Tool Orchestrator**: Execute and manage external tools/functions

#### 2. Memory System (Hybrid Approach)

- **Working Memory**: Current conversation context
- **Session Memory**: Recent conversation history
- **Long-term Memory**: RAG-based knowledge storage
- **Conversation Analysis**: Extract patterns, preferences, personality traits
- **Knowledge Forgetting**: Automatic cleanup of outdated information

#### 3. Data Pipeline (Future Expansion)

- **Text Processing**: Documents, websites, chat history
- **Media Processing**: Videos (transcription), images (OCR/vision)
- **Knowledge Extraction**: Convert raw data into searchable knowledge
- **Storage**: SQLite + vector database (Chroma/LanceDB)

---

## Model Requirements

- **Tool-Capable**: Support for function calling and external tool execution
- **Local-Runnable**: Efficient enough for local hardware
- **Capable**: Strong reasoning and conversation abilities

---

## Self-Awareness Definition

The assistant should demonstrate:

- **Memory Continuity**: Remember and reference past conversations
- **Capability Awareness**: Understand its own abilities and limitations
- **Personality Consistency**: Maintain character traits across sessions
- **Learning Recognition**: Acknowledge when it has learned something new

---

## Technical Considerations

### Performance

- Efficient vector search for large knowledge bases
- Streaming responses for better user experience
- Context window management for long conversations
- Resource usage optimization for local execution
- Think block filtering preserves full context while hiding reasoning from display

### Security

- Sandboxes code execution for safety
- Secure storage of sensitive information
- Safe handling of external API keys

### Scalability

- Modular architecture for easy feature addition
- Plugin system for third-party tools
- Configurable resource limits
- Backup and restore capabilities

---

## Card System (Personas & System Prompts)

Cards define assistant personas, prompts, system behavior, and (later) allowed tools.

Format: Markdown with optional front matter (YAML or TOML). If front matter is
absent, entire file (minus leading heading) is the system prompt.

Example (`researcher.md`):

```markdown
---
name: Researcher
description: Focused, methodical analyst
tags: [research, analysis]
---

You are a focused research assistant. Provide concise, sourced summaries when possible.
```

Resolution order (first match wins):

1. Explicit `-card` flag (future)
2. Default card configured (future config)
3. Built-in fallback system prompt

Tool permission gating is deferred until tools beyond web search are added.

---

## Configuration Precedence

Flags override environment variables, which override internal defaults.

| Concern | Flag     | Env             | Default (MVP)         |
| ------- | -------- | --------------- | --------------------- |
| Model   | `-model` | `DEFAULT_MODEL` | (required if not set) |

- Recoverable LLM failures (transport, non-2xx, decode) surface as system
  messages so sessions can continue without exiting.

---

## Exit Codes

| Code | Meaning                                         |
| ---- | ----------------------------------------------- |
| 0    | Success                                         |
| 1    | Generic runtime error                           |
| 2    | Invalid CLI usage / config error                |
| 3    | Model unavailable (host down, model not pulled) |
| 4    | Tool execution rejected or failed critically    |
| 5    | Context cancellation / timeout                  |

Codes may expand; backward compatibility will be maintained after first tag.

---

## Logging Strategy

- All logs written to `~/.ghost/ghost.log` at DEBUG level (always enabled).
- Logs use JSON formatting for structured output with caller information and timestamps.
- Model/token output to stdout (enables piping).
- stderr reserved for critical startup failures only.
- Explicit debug tooling (e.g., spew dumps) may write structured data to
  stdout when gated behind a developer-only flag.
- All operations accept `context.Context` for cancellation and trace correlation.
- Avoid panics outside `main`; return errors with `%w` for wrapping.

Use sentinel errors (package-level variables) for robust error matching and
propagation. Wrap sentinel errors with `%w` when returning from functions,
and prefer `errors.Is` for error checks in tests and consumers.

**Error Message Guidelines:**

- Use consistent "failed to..." pattern for sentinel error messages
- Domain-based sentinels (e.g., `ErrConfig`, `ErrLogging`) over cause-based
- Only create sentinels for errors that callers need to handle differently

**Error Wrapping Best Practices:**

- Always use `%w: %w` when wrapping an error with a sentinel error
- The first `%w` is for the sentinel, the second `%w` is for the underlying error
- This preserves the full error chain for `errors.Is()` and `errors.Unwrap()`
- Never use `%s` for error arguments as it breaks the error chain

Example:

```go
// Sentinel error definitions
var (
    ErrConfig  = errors.New("failed to bind config")
    ErrLogging = errors.New("failed to setup logging")
)

// Wrapping with sentinel - CORRECT
if err != nil {
    return fmt.Errorf("%w: %w", ErrConfig, err)
}

// Wrapping with sentinel - INCORRECT (breaks error chain)
if err != nil {
    return fmt.Errorf("%w: %s", ErrConfig, err)  // Don't use %s!
}

// Wrapping with context and status codes
if statusCode/100 != 2 {
    return fmt.Errorf("%w: status=%d %s: %w", ErrNon2xxResponse,
        statusCode, http.StatusText(statusCode), err)
}

// Checking in consumer
if errors.Is(err, ErrConfig) {
    // handle configuration error specifically
}
```

Log levels (guideline):

- DEBUG: token streaming internals, detailed state information
- INFO: start/end of requests, selected model, tool invocations
- WARN: transient recoverable issues (recoverable network errors)
- ERROR: user-visible failures or abort conditions

**Security Note:** Logs never contain user input content (queries, responses, arguments)
to prevent sensitive data leakage. Only metadata (lengths, counts, status codes)
 is logged.

---

## Testing & Quality

- Each internal package defines an interface to enable mocking in dependents
  (e.g., `llm.Client`).
- Separate test cases for a function using `test.Run()`.
- Run tests in parallel when possible using `t.Parallel()`.
- Use sentinel errors and `errors.Is` for error assertions in tests, rather
  than string matching.
- Avoid magic strings by hoisting shared literals (messages, prompts, keys)
  into constants shared across code and tests.
- Run: `go test ./...`; optional: `go vet ./...`; later: integrate `golangci-lint`.
- Race checks: `go test -race` (periodic / CI optional early on).

## Tool Execution

- The agent decides if it needs to run a tool.
- Rejections return an error mapped to exit code 4 if fatal.

---

## Concurrency & Streaming Guidelines

- A single streaming response pipeline per request; use channels for token
  delivery.
- Cancel on context done; ensure goroutines exit (no leaks).
- Backpressure: token writer checks context and downstream errors; do not
  buffer unbounded.

## Style & Conventions (Summary)

- Package boundaries mirror architecture (`internal/llm`, `internal/tools`, etc.).
- `context.Context` as first parameter after receiver.
- Errors wrapped with `%w`; user-facing messages composed at the edge (CLI layer).
- Prefer standard library; introduce third-party deps only with justification.
