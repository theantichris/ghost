# Spec

**Ghost** is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. It is designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_, _Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always on AI companion into a terminal-first experience.

Capabilities should include research, web searching, helping with code, generating images, executing tasks, setting up reminders, and chatting.

---

## MVP Scope (Initial Deliverable)

The first tagged release (MVP) intentionally limits scope for fast iteration and stability:

- CLI only (no TUI yet) at `./cmd/assistant`
- Streaming chat with a single Ollama model (flag or env configured)
- Basic web search tool (agent can decide to use tools without confirmation)
- Simple tool execution framework (just enough to support web search & future expansion)
- Basic logging (structured via `slog`) to stderr; model output to stdout
- No persistent long‑term memory (session context only in-process)
- No embedding / vector store yet (deferred)
- Persona / Prompt system loading (Markdown cards)

Everything else (TUI, RAG, multi-model abstraction, automation scheduling) is expressly out of MVP scope.

Prompts and character personalities will be loaded via a **Card system**.

Initially Ghost will use the **Ollama API** as the backend to access LLMs.

Initial interaction will be through **CLI**; a **TUI** is planned post‑MVP.

---

## Technical Architecture

### Core Principles

- **Local Execution**: Core AI processing runs locally via Ollama
- **Internet-Enabled**: External services (web search, APIs) allowed for enhanced capabilities
- **Modular Design**: Clean separation between engine, memory, tools, and interfaces
- **Strong Separation of Concerns**: Technical designs must enforce clear boundaries between components and responsibilities to keep the project easy to maintain and extend.

### System Components

#### 1. Core Engine

- **LLM Client**: Ollama API integration with tool support
- **Conversation Manager**: Handle chat flow, context windows, streaming
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

### Security

- Sandboxed code execution for safety
- Secure storage of sensitive information
- Access control for different assistants
- Safe handling of external API keys

### Scalability

- Modular architecture for easy feature addition
- Plugin system for third-party tools
- Configurable resource limits
- Backup and restore capabilities

---

## Card System (Personas & System Prompts)

Cards define assistant personas, prompts, system behavior, and (later) allowed tools. For MVP only system prompt + name are required.

Format: Markdown with optional front matter (YAML). If front matter absent, entire file (minus leading heading) is the system prompt.

Example (`cards/researcher.md`):

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

| Concern | Flag             | Env                   | Default (MVP)         |
| ------- | ---------------- | --------------------- | --------------------- |
| Model   | `-model`         | `OLLAMA_MODEL`        | (required if not set) |
| Stream  | `-stream` (bool) | N/A                   | true                  |
| Card    | (future `-card`) | `GHOST_CARD` (future) | built‑in basic        |

Config file (TOML/YAML) deliberately deferred.

---

## Exit Codes (MVP Baseline)

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

- Use `log/slog` with JSON handler when `LOG_FORMAT=json`, otherwise text.
- Logs to stderr; model/token output to stdout (enables piping).
- All operations accept `context.Context` for cancellation and trace correlation (future: add request IDs).
- Avoid panics outside `main`; return errors with `%w` for wrapping.

Use sentinel errors (package-level variables, e.g., `var ErrModelEmpty = errors.New("model cannot be empty")`) for robust error matching and propagation. Wrap sentinel errors with `%w` when returning from functions, and prefer `errors.Is` for error checks in tests and consumers.

Log levels (guideline):

- DEBUG: token streaming internals (suppressed by default)
- INFO: start/end of requests, selected model, tool invocations
- WARN: transient recoverable issues (retryable network errors)
- ERROR: user-visible failures or abort conditions

---

## Testing & Quality

- Each internal package defines an interface to enable mocking in dependents (e.g., `llm.Client`).
- Provide a lightweight fake Ollama client for deterministic tests (no network).
- Golden tests (commit stable expected outputs) only at boundaries that are stable and low-churn (e.g., request serialization, tool manifest formatting).
  Use sentinel errors and `errors.Is` for error assertions in tests, rather than string matching. Golden tests (commit stable expected outputs) only at boundaries that are stable and low-churn (e.g., request serialization, tool manifest formatting).
- Run: `go test ./...`; optional: `go vet ./...`; later: integrate `golangci-lint`.
- Race checks: `go test -race` (periodic / CI optional early on).

---

## Memory System Phases

| Phase   | Scope                   | Persistence                | Notes                      |
| ------- | ----------------------- | -------------------------- | -------------------------- |
| 0 (MVP) | In-process context only | None                       | Simple slice of messages   |
| 1       | Session memory          | File (JSON)                | Rotated per session ID     |
| 2       | Long-term embeddings    | SQLite + future vector ext | Lightweight indexing       |
| 3       | Conversation analysis   | Derived metadata           | Preference extraction      |
| 4       | Knowledge pruning       | Background job             | Age + relevance thresholds |

RAG and embeddings start no earlier than Phase 2.

---

## Tool Execution (Interactive Safety)

MVP includes a web search tool (implementation detail TBD). Rules:

- The agent decides if it needs to run a tool.
- Rejections return an error mapped to exit code 4 if fatal.
- Future: policy config to auto-approve certain tools.

Execution boundaries (planned): no arbitrary file writes until explicit opt‑in tool added.

---

## Concurrency & Streaming Guidelines

- A single streaming response pipeline per request; use channels for token delivery.
- Cancel on context done; ensure goroutines exit (no leaks).
- Backpressure: token writer checks context and downstream errors; do not buffer unbounded.
- Timeouts configurable later; MVP may rely on user Ctrl+C.

---

## Future (Deferred) Decisions

| Topic                                    | Status                 |
| ---------------------------------------- | ---------------------- |
| Vector store (Chroma/LanceDB/SQLite FTS) | Deferred until Phase 2 |
| Config file                              | Deferred               |
| Multi-backend LLM abstraction            | Out of scope MVP       |
| Tool policy DSL                          | Deferred               |
| TUI (Bubble Tea)                         | Post-MVP milestone     |

---

## Minimal Threat Model

Attack surfaces (early):

- Web search tool (external HTTP requests)
- LLM prompt injection via user input or retrieved content
- Future tool execution proposals

Initial mitigations:

- Strict separation between system prompt and user content
- No shell execution or file system writes in MVP
- Environment variables only read; never logged verbatim

Deferred mitigations:

- Output sanitization for rendered content
- Sandboxed execution for code tools
- API key scoping & encryption at rest

---

## Style & Conventions (Summary)

- Package boundaries mirror architecture (`internal/llm`, `internal/tools`, etc.).
- `context.Context` as first parameter after receiver.
- Errors wrapped with `%w`; user-facing messages composed at the edge (CLI layer).
- Prefer standard library; introduce third-party deps only with justification.

---

## Open Questions / TBD

1. Exact web search provider & implementation details
2. Card discovery order customization (beyond MVP defaults)
3. Structured tracing / metrics approach (if any) post-MVP

These remain intentionally unspecified to preserve flexibility.
