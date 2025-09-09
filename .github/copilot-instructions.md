# GitHub Copilot Instructions for assistant-go

These instructions narrow Copilot’s behavior for this repository. Always align generated code and docs with:

- SPEC: ../SPEC.md
- Roadmap: ../ROADMAP.md

Principles

- Local-first, privacy-first. Default to local execution via Ollama; avoid cloud dependencies and telemetry by default.
- Respect user-provided API keys securely.
- Cross-platform development (Windows, macOS, Linux). Ensure paths, shells, and filesystems are handled portably.
- Small, composable components with clear interfaces; predictable flags and output.

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
