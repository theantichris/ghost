# Agent Guidelines for Ghost

Quick reference for automated coding agents and developers working on Ghost.

See [SPEC.md](/SPEC.md) for full technical architecture, error handling patterns,
 logging strategy,
 and system design.

## Build & Test Commands

- **Build**: `go build -v ./...`
- **Test all**: `go test -v ./...`
- **Test single package**: `go test -v ./cmd` or `go test -v ./internal/llm`
- **Run single test**: `go test -v -run TestFunctionName ./path/to/package`
- **Lint**: `golangci-lint run` (via pre-commit hook)
- **Format**: `go fmt ./...` (automatically via pre-commit)
- **Release**: `git tag vX.Y.Z && git push origin vX.Y.Z` (triggers
 GoReleaser via GitHub Actions)

## Code Style

- **Imports**: Standard library first, then third-party, then internal packages
 (blank line separation)
- **Error handling**: Define sentinel errors as `var Err... = errors.New("...")`
 at package level; wrap errors with `fmt.Errorf("%w: %w", ErrSentinel, err)`
- **Interfaces**: Define minimal interfaces in consumer packages (e.g., `LLMClient`
 in `internal/llm/client.go`)
- **Naming**: Use camelCase for unexported, PascalCase for exported; descriptive
 names (e.g., `llmClient`, `chatHistory`)
- **Testing**: Table-driven tests with `t.Run()` for subtests; use `t.Parallel()`
 for parallel tests; mock interfaces for dependencies
- **Comments**: Only add comments for exported functions/types or complex logic
- **Logging**: Use `charmbracelet/log` for structured logging with key-value pairs;
 log to stderr by default for pipeline friendliness. Never log secrets or sensitive
 information; redact as needed
- **Configuration**: Viper for config management; support env vars, flags, and
 TOML config files
- **Types**: Prefer concrete types. If an empty interface is unavoidable, use `any`
 instead of `interface{}`
- **Constants**: Define as typed constants in blocks at package level

## File Organization

Ghost follows a modular file organization pattern within packages to maintain
 clear separation of concerns:

### `cmd/` Package Structure

The `cmd/` package is organized by responsibility rather than having
 monolithic command files:

- **`ask.go`** - Ask command definition and orchestration logic
- **`root.go`** - Root command setup, configuration initialization, and logging setup
- **`llm.go`** - LLM client initialization and shared constants (e.g., `systemPrompt`)
- **`input.go`** - Input handling utilities (`readPipedInput`)
- **`format.go`** - Output formatting utilities (`stripThinkBlock`)
- **`errors.go`** - Centralized sentinel error definitions for the cmd package

### When to Create New Files

Follow these guidelines when deciding whether to create a new file or add to
 an existing one:

- **Create a new file when:**
  - A function is reusable across multiple commands (e.g., `readPipedInput
   used by ask and future chat commands)
  - A file exceeds ~300 lines and has distinct logical sections
  - You're adding a new command (e.g., `chat.go`, `search.go`)
  - You have 5+ related utility functions that form a cohesive group

- **Add to existing file when:**
  - The function is only used by one command
  - The file is under 200 lines and cohesive
  - The function is tightly coupled to existing code in the file

- **File naming conventions:**
  - Use concrete, descriptive names: `input.go`, `format.go`, `client.go`
  - Avoid generic names like `utils.go`, `helpers.go`, or `common.go`
  - Name command files after the command: `ask.go`, `chat.go`

### YAGNI Principle

Avoid premature abstraction. Don't create `internal/` packages until you have
 concrete reuse across multiple top-level packages. For example, `cmd/input
  go` should only move to `internal/io` when packages outside `cmd/` need the
   same functionality.

## Pre-commit Hooks

The following checks are run on every commit via pre-commit hooks: `go fmt`,
 `go mod tidy`, `go test`, `golangci-lint`, markdown linting, and spell checking.

> **Note:** To enable these checks locally, you must install [pre-commit](https://pre-commit.com/)
 and run `pre-commit install` in your repository root. See the `.pre-commit-config.yaml`
  file for the list of configured hooks.

## Release Process

Ghost uses [GoReleaser](https://goreleaser.com/) for automated releases via
 GitHub Actions.

### Creating a Release

1. Tag the commit with a semantic version: `git tag vX.Y.Z`
2. Push the tag: `git push origin vX.Y.Z`
3. GitHub Actions will automatically:
   - Build binaries for Linux, macOS, and Windows
   - Create archives (tar.gz for Unix, zip for Windows)
   - Generate a changelog from commits
   - Create a GitHub release with all artifacts

### Configuration

- **GoReleaser config**: `.goreleaser.yaml` in project root
- **GitHub workflow**: `.github/workflows/release.yml`
- Builds are CGO-disabled for maximum portability
- Archives follow OS/Arch naming conventions compatible with `uname`
