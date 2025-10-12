# Agent Guidelines for Ghost

Quick reference for automated coding agents and developers working on Ghost.

See [SPEC.md](/SPEC.md) for full technical architecture, error handling patterns,
 logging strategy,
 and system design.

## Build & Test Commands

- **Build**: `go build -v ./...`
- **Test all**: `go test -v ./...`
- **Test single package**: `go test -v ./cmd` or `go test -v ./internal/llm`
  or `go test -v ./internal/stdio`
- **Run single test**: `go test -v -run TestFunctionName ./path/to/package`
- **Lint**: `golangci-lint run` (via pre-commit hook)
- **Format**: `go fmt ./...` (automatically via pre-commit)
- **Release**: `git tag vX.Y.Z && git push origin vX.Y.Z` (triggers
 GoReleaser via GitHub Actions)

## Streaming Implementation Pattern

The chat command uses BubbleTea's channel-based streaming pattern for
real-time token display:

1. **Command Structure**: `sendChatRequest()` returns `tea.Cmd` (not `tea.Msg`)
2. **Channel Creation**: Create a channel in the command's goroutine:
   `sub := make(chan tea.Msg)`
3. **Token Delivery**: Send each token as
   `streamingChunkMsg{content: token, sub: sub}` including the channel reference
4. **Continuous Listening**: Use `waitForActivity(msg.sub)` in `Update()` to
   return a command that waits for the next message
5. **Completion**: Send `streamCompleteMsg` when streaming finishes, close
   the channel
6. **Viewport Rendering**: Include `model.currentMsg` in `wordwrap()` to
   display streaming tokens in real-time

**Critical Configuration**: Timeout must have a default value via
`viper.SetDefault("timeout", 2*time.Minute)` in `initConfig()`. Without this,
`viper.GetDuration("timeout")` returns 0, causing `context.WithTimeout` to
create already-expired contexts that fail immediately.

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
 for parallel tests; mock interfaces for dependencies using dependency injection
 (e.g., injecting `llm.LLMClient` into command structs). In test assertions, use
 explicit variable names with `actual` and `expected` prefixes (e.g., `actualOutput`,
 `expectedOutput`, `actualTokens`, `expectedTokens`) to make comparisons clear.
 Add a blank line before assertion blocks to improve readability. All pre-commit
 hooks must pass before committing changes. Current test coverage: cmd 64.5%,
 internal/llm 69.9%
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

- **`ask.go`** - Ask command with LLM client dependency injection
- **`chat.go`** - Interactive chat command with TUI interface and history management
- **`root.go`** - Root command setup, configuration initialization, and logging setup
- **`llm.go`** - LLM client initialization and shared constants (e.g., `systemPrompt`)
- **`errors.go`** - Centralized sentinel error definitions for the cmd package

### `internal/stdio/` Package Structure

Standard input/output handling utilities used by command implementations:

- **`input.go`** - Input handling utilities (`InputReader`, piped input
  detection, user input scanning)
- **`output.go`** - Output stream processing with think block filtering (`OutputWriter`)
- **`errors.go`** - I/O-specific sentinel errors (`ErrIO`, `ErrInputEmpty`)

### `internal/tui/` Package Structure

Terminal user interface implementation for interactive chat:

- **`model.go`** - BubbleTea model implementation with streaming message handling,
  viewport rendering, and LLM integration. Implements channel-based streaming
  pattern where `sendChatRequest()` returns `tea.Cmd`, tokens are sent as
  `streamingChunkMsg` with channel reference, and `waitForActivity()` continues
  listening for streaming messages. Exit handling distinguishes between immediate
  exit (Ctrl+D/Ctrl+C) and graceful exit routine (`/bye`, `/exit`) which displays
  a goodbye message and awaits user keypress via the `awaitingExit` state.
- **`run.go`** - TUI program entry point
- **`errors.go`** - TUI-specific sentinel errors (`ErrLLMClientInit`, `ErrLLMRequest`)

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

### Agent Workflow

**IMPORTANT**: Whenever you create or edit any file, you MUST run the appropriate
linter and fix any issues before considering the task complete:

- **Go files**: Run `go fmt` and `golangci-lint run` on the modified files
- **Markdown files**: Run `pre-commit run markdownlint --files <filename>`
  to check and auto-fix markdown issues
- **All files**: Consider running `pre-commit run --files <filename>` to
  run all applicable hooks

Do not wait for the user to ask you to lint. This should be automatic for
every file you touch.

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
