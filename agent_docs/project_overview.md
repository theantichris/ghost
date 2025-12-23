# Project Overview – Ghost

## What is Ghost?

Ghost is a terminal-based AI assistant CLI tool that acts as a local AI
companion. It connects to Ollama to provide:

- Text-based chat interactions
- Piped input processing (up to 10 MB)
- Image analysis with vision models
- Health diagnostics for the Ollama connection

## Architecture

### Entry Point

`main.go`:

- Initializes structured logger to `~/.config/ghost/ghost.log`
- Creates 5-minute timeout context
- Calls `cmd.Run()` and extracts exit code on error

### CLI Layer (`internal/cmd/`)

**`main.go`**: Root command setup

- Uses `urfave/cli/v3` for command structure
- Loads optional TOML config from `~/.config/ghost/config.toml`
- Defines flags with TOML source chain for config file fallback
- `before` hook: Initializes `llm.LLMClient` and adds to metadata
- Root action: Handles prompt + piped input + images
- Subcommands: `health`

**`health.go`**: Diagnostics subcommand

- Displays system config (host, models, prompts)
- Tests Ollama API connectivity
- Validates that both chat and vision models are available

**`config.go`**: Config file loading

- Returns `altsrc.StringSourcer` for TOML integration
- Gracefully handles missing config file (not an error)

**`errors.go`**: Command-specific error definitions

- `ErrNoPrompt`, `ErrInput`, `ErrConfigFile`
- All wrapped with `exitcode` exit codes

### LLM Layer (`internal/llm/`)

**`client.go`**: Interface definition

```go
type LLMClient interface {
    Generate(ctx, systemPrompt, userPrompt string, images []string) (string, error)
    Version(ctx) (string, error)
    Show(ctx, model string) error
}
```

**`ollama.go`**: Ollama API implementation

- Uses `carlmjohnson/requests` for HTTP calls
- Endpoints: `/api/generate`, `/api/version`, `/api/show`
- Streaming disabled for simplicity
- Handles both default model and vision model

**`mock.go`**: Test mock

- `MockLLMClient` with function fields for test stubbing
- Allows per-method behavior customization

**`errors.go`**: LLM-specific errors

- `ErrOllama`, `ErrModelNotFound`, `ErrNoHostURL`, etc.
- All wrapped with appropriate exit codes

### Exit Code Layer (`internal/exitcode/`)

**`exitcode.go`**: sysexits.h-style exit codes

- `ExitCode` type with constants (`ExUsage`, `ExConfig`, `ExUnavailable`, etc.)
- `Error` type wraps errors with exit codes
- `GetExitCode(err)` extracts code from error chain
- Preserves error wrapping for `errors.Is/As`

### Dual Model Flow

1. User provides prompt + optional images
2. If images present:
   - Call vision model with `--vision-prompt` and images
   - Append vision model response to user prompt
3. Call chat model with final prompt
4. Return result to stdout

This is **not** a single multi-modal call. Vision analysis happens first, then
text generation.

### Configuration

**TOML structure**:

```toml
host = "http://localhost:11434"
model = "llama3.1:8b"
system = "System prompt override"

[vision]
model = "qwen2.5vl:7b"
system_prompt = "Vision system override"
prompt = "Analyze the attached image(s)"
```

**Priority**: CLI flags > TOML config > flag defaults

**Loading**: Uses `cli-altsrc/v3` with `toml.TOML("key", configFile)` sources

### Logging

- Logger: `charmbracelet/log` with JSON formatter
- Output: `~/.config/ghost/ghost.log`
- Options: `ReportCaller: true`, `ReportTimestamp: true`, `Level: DebugLevel`
- Pass logger via metadata or function parameters
- Use structured fields: `logger.Debug("message", "key", value)`

### Testing

- **Unit tests**: Table-driven with `goldie/v2` golden files
- **Golden files**: In `internal/cmd/testdata/TestName/subtest_name.golden`
- **Mocking**: `llm.MockLLMClient` for LLM operations
- **E2E tests**: `e2e_test.go` builds binary and executes (requires ollama)
- **Update snapshots**: `go test -update`

### Key Files

- `main.go`: Application entry point
- `internal/cmd/main.go`: CLI setup and root action
- `internal/cmd/health.go`: Health check subcommand
- `internal/llm/ollama.go`: Ollama API client
- `internal/exitcode/exitcode.go`: Exit code handling
- `e2e_test.go`: End-to-end tests
- `.pre-commit-config.yaml`: Pre-commit hooks (go-fmt, golangci-lint, etc.)
- `.golangci.yml`: Linter config
- `.harper-dictionary.txt`: Custom spell check dictionary
