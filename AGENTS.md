# AGENTS.md

AI agent guidance for the Ghost codebase.

## Project Overview

**Ghost** is a local AI assistant CLI tool built with Go, powered by Ollama.
Inspired by cyberpunk media (Shadowrun, Cyberpunk 2077, The Matrix).

## Code Conventions

**Style**:

- Standard Go formatting (enforced by pre-commit)
- Wrap errors with `fmt.Errorf("%w", err)`
- Follow Go naming conventions (exported vs unexported)
- Comment struct fields and exported types

**Testing**:

- Use table-driven tests pattern
- Use `errors.Is()` for error comparison
- Use `t.Fatalf()` for unexpected errors, `t.Errorf()` for assertions

**Commit Messages**: Conventional commits (`feat:`, `fix:`, `refactor:`,
`test:`, `docs:`)

## Design Principles

- **Keep it simple**: Single-file structure unless strong reason to split
- **Cyberpunk aesthetic**: Match tone in user-facing messages
- **CLI-first**: Prioritize terminal experience
