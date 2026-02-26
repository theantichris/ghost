# CLAUDE.md

## Project Overview

Ghost is a command line AI assistant written in Go and powered by Ollama, designed
with a cyberpunk aesthetic inspired by Shadowrun, Cyberpunk 2077, and The Matrix.

## Configuration

Priority (highest to lowest):

1. Command-line flags
2. Environment variables (prefixed with `GHOST_`, dots/hyphens replaced with `*`)
3. Config file (`~/.config/ghost/config.toml`)

Nested config keys use dot notation in TOML, hyphens in flags, and `*` in env
vars (e.g., `vision.model` / `--vision-model` / `GHOST_VISION*MODEL`).

## Code Conventions

- Standard Go formatting (enforced by pre-commit)
- Wrap errors with `fmt.Errorf("%w", err)` for proper error chains
- Cyberpunk aesthetic in user-facing messages (e.g., "neural link", "data stream")
- UI glyphs: `theme.GlyphInfo` (󱙝) and `theme.GlyphError` (󱙜)
- Table-driven tests, one test function per code function
  (e.g., `TestChatModel_HandleCommandMode` tests `handleCommandMode`)
- Test file naming mirrors source files (e.g., `chat_command_test.go`)
- Use `errors.Is()` for error comparison
- Conventional commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`,
 `perf:`, `build:`, `ci:`, `style:`)

## Design Principles

- **Keep it simple**: Single file structure per package unless strong reason to split
- **Cyberpunk aesthetic**: Match tone in user facing messages and error messages
- **TUI first**: Prioritize terminal experience with proper TTY detection
