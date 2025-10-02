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

## Pre-commit Hooks

All commits automatically run: `go fmt`, `go mod tidy`, `go test`, `golangci-lint`,
 markdown linting, and spell checking.
