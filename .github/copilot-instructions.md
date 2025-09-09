# GitHub Copilot Instructions for assistant-go

These instructions narrow Copilot’s behavior for this repository. Always align generated code and docs with:

- SPEC: ../SPEC.md
- Roadmap: ../ROADMAP.md

Principles

- Local-first, privacy-first. Default to local execution via Ollama; avoid cloud dependencies and telemetry by default.
- Respect user-provided API keys securely.
- Cross-platform development (Windows, macOS, Linux). Ensure paths, shells, and filesystems are handled portably.
- Small, composable components with clear interfaces; predictable flags and output.
- Enforce strong separation of concerns in technical designs to keep the project easy to maintain and extend.

Project structure (recommended)

- cmd/assistant/main.go (CLI entrypoint)
- internal/llm (Ollama client, streaming)
- internal/tui (Bubble Tea scaffolding for chat view)
- internal/tools (tool framework and built-ins)
- internal/memory (session/history, RAG integration later)
- internal/config (flag/env/config precedence)
- pkg/... only if a stable external API is needed

CLI expectations (planned)

- Flags: -model, -stream; env: OLLAMA_MODEL, OLLAMA_HOST; precedence flag > env > default.
- Input via args or stdin; detect TTY vs stdin; graceful cancel via context.
- Stream tokens to stdout when -stream is true; meaningful exit codes and errors.

TUI expectations (planned)

- Bubble Tea single-window chat view with streaming render, input line, scrollback, and basic keymap.

Tools and safety (planned)

- Minimal tool interface for function-calling; built-ins may include read-only FS info and opt-in shell exec.

Memory and RAG (future)

- Session and long-term memory with a simple index (filesystem/SQLite) and local embeddings when introduced.

Quality and CI

- Use contexts/timeouts; avoid global state; write unit tests; consider golden tests where stable.
- Keep concurrency safe and cancelable; avoid blocking the UI.
- Ensure `go build ./...` and `go test ./...` pass; prefer `go vet` and `-race` where applicable.

Documentation

- Update README.md, LLMS.md, AGENTS.md, and WARP.md when public surfaces change.
- Do not add telemetry or hidden network calls.

Out of scope (initially)

- Cloud-hosted LLM backends by default; invasive system control; any default-on telemetry.

---

## Implementation Standards (Augmented)

These additions reflect clarified project preferences (solo dev, MVP first tag).

### Architecture & Packages

- Keep features behind small interfaces inside `internal/` packages (e.g., `llm.Client`, `tools.Executor`).
- Avoid cyclic dependencies; packages depend inward (CLI -> internal/\*).
- Add new external-facing APIs only if they are stable; otherwise keep them internal.

### Interfaces & Testing

- Every package that is consumed by another defines an interface + a mock/fake in a `_test.go` file or a dedicated `testutil` subpackage.
- Fakes should avoid network calls (e.g., fake Ollama client returns scripted token streams).
- Wrap errors with `%w` for reliable `errors.Is`/`errors.As` use.
- Golden tests only for stable serialization boundaries (request payloads, card parsing output).

### Logging

- Use `log/slog` exclusively; no ad-hoc fmt logging in library code.
- Logs go to stderr; streamed model tokens go exclusively to stdout.
- Provide helper constructor for logger with text vs JSON selection via env (`LOG_FORMAT`).

### Context & Cancellation

- Every public method that can block takes `context.Context` as first parameter (after receiver) and respects early cancellation.
- Ensure all goroutines terminate on context cancellation (no leaks). When launching goroutines, add comments explaining purpose.

### Error Handling

- Do not panic except in `main` for truly unrecoverable initialization failures.
- Wrap lower-level errors (`fmt.Errorf("describe: %w", err)`) at boundaries; avoid repeated wrapping at same layer.
- User-facing phrasing is assembled at CLI boundary; internals keep errors factual.

### Exit Codes (Reference)

| Code | Meaning                        |
| ---- | ------------------------------ |
| 0    | Success                        |
| 1    | Runtime error                  |
| 2    | Invalid usage / configuration  |
| 3    | Model unavailable / host issue |
| 4    | Tool failure / denied          |
| 5    | Canceled / timeout             |

### Tools (MVP Constraint)

- Only a web search tool is planned initially.
- No shell execution or filesystem mutation tools until explicitly introduced.

### Cards (Personas)

- Markdown files with optional YAML front matter in `cards/` directory.
- Fallback persona lives in code; card selection future flag (`-card`) or env (`GHOST_CARD`).

### Dependency Policy

- Favor standard library; introduce third-party dependencies only with clear advantage (document rationale in PR description/commit message if added).
- No hidden telemetry or background network calls.

### Linting & Quality

- Plan to adopt `golangci-lint`; until configured, manually run `go vet` and consider `-race` in CI.
- Keep functions short and cohesive; prefer explicit error handling over cleverness.

### Platform Compatibility

- Use `filepath` not hard-coded separators.
- Avoid assumptions about ANSI support; TUI will negotiate later.

### Security Posture (Early)

- No credential logging; redact or omit.
- Avoid executing user-provided code (no eval) in MVP.

### Concurrency

- Document any goroutine spawn with a comment including: purpose, ownership, cancellation path.
- Avoid unbounded buffering; prefer streaming channels with backpressure (small buffers or none).

### Future Deferrals (Do Not Implement Prematurely)

- Vector store and embedding pipeline.
- Multi-backend LLM abstraction beyond Ollama.
- Persistent memory (beyond in-process session state) until memory phase begins.
- Config file parsing (keep to flags + env for now).

### Commit / PR Style (When Applicable)

- Imperative tense: "Add streaming client" / "Fix card parsing error".
- Reference spec section if clarifying or extending architecture.

### Style Summary

- `context` first, errors last return value.
- Keep exported surface minimal; export only what tests or CLI need.
- Avoid generics unless they simplify code substantially.

---
