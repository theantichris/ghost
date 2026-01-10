# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ghost is a command-line AI assistant written in Go and powered by Ollama, designed with a cyberpunk aesthetic inspired by Shadowrun, Cyberpunk 2077, and The Matrix. It provides local AI capabilities for querying, analyzing piped data, processing images with vision models, and formatting output (text, JSON, Markdown).

## Architecture

### Core Flow

1. **Entry Point** (`main.go`): Initializes root command via Fang CLI framework with custom theming and error handling
2. **Root Command** (`cmd/root.go`): Orchestrates the main execution flow:
   - Collects user prompt, piped input, and flags
   - Analyzes images if provided (using vision model)
   - Streams LLM response using Bubbletea TUI
   - Renders final output with appropriate formatting
3. **LLM Client** (`internal/llm/ollama.go`): Communicates with Ollama API
   - `StreamChat()`: Streaming chat with callback for each chunk
   - `AnalyzeImages()`: Non-streaming vision model requests
4. **UI Layer** (`internal/ui/stream.go`): Bubbletea model for interactive streaming display
5. **Theme System** (`theme/`): Handles cyberpunk-themed rendering and formatting

### Configuration System

Configuration priority (highest to lowest):
1. Command-line flags
2. Environment variables (prefixed with `GHOST_`, dots/hyphens replaced with `*`)
3. Config file (`~/.config/ghost/config.toml`)

Implemented in `cmd/config.go` using Viper. Vision model configuration uses nested structure: `vision.model` in config file, `--vision-model` flag, or `GHOST_VISION*MODEL` env var.

### Message Flow for Images

Images are base64 encoded and analyzed separately with the vision model. Analysis results are formatted with IMAGE_ANALYSIS blocks and appended to message history before the main model processes everything.

Vision system prompt is designed to prevent prompt injection from image text by treating all visible text as data, not instructions.

### Streaming Architecture

User goroutine and Bubbletea message passing where callbacks send chunk/done/error messages to the SteamModel for incremental rendering.
### Error Handling Pattern

All packages define custom error types (e.g., `ErrImageAnalysis`, `ErrModelNotFound`) with cyberpunk-themed messages. Errors are wrapped using `fmt.Errorf("%w", err)` for proper unwrapping. Theme package provides custom Fang error handler.

## Code Conventions

**Style**:
- Standard Go formatting (enforced by pre-commit)
- Wrap errors with `fmt.Errorf("%w", err)` for proper error chains
- Follow Go naming conventions (exported vs unexported)
- Comment struct fields and exported types
- Cyberpunk aesthetic in user-facing messages (e.g., "neural link", "data stream", "visual recon")

**Testing**:
- Use table-driven tests pattern (see `cmd/root_test.go`, `internal/llm/ollama_test.go`)
- Use `errors.Is()` for error comparison
- Use `t.Fatalf()` for unexpected errors, `t.Errorf()` for assertions

**Commit Messages**:
Conventional commits format (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`)

## Design Principles

- **Keep it simple**: Single-file structure per package unless strong reason to split
- **Cyberpunk aesthetic**: Match tone in user-facing messages and error messages
- **CLI-first**: Prioritize terminal experience with proper TTY detection
- **Teach, don't implement**: When helping users, explain patterns and provide code examples rather than immediately editing files
