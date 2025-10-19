# Agent Guidelines for Ghost

## Build/Test/Lint Commands

- **Build**: `go build -v ./...`
- **Test all**: `go test -v ./...`
- **Test single**: `go test -v ./internal/cmd -run TestHandleLLMRequest`
- **Lint**: `golangci-lint run` (via pre-commit hooks)
- **Format**: `go fmt ./...` (enforced via pre-commit)
- **Tidy**: `go mod tidy`

## Code Style

- **Go version**: 1.24.2
- **Imports**: Standard library first, then third-party (e.g.,
  charmbracelet/log, urfave/cli/v3, carlmjohnson/requests)
- **Errors**: Define sentinel errors as package-level vars wrapped with exit
  codes using `exitcode.New(errors.New("message"), exitcode.ExCode)`. Wrap
  errors in functions with `fmt.Errorf("%w: %w", ErrParent, err)`. Extract exit
  codes at application boundary with `exitcode.GetExitCode(err)`.
- **Exit Codes**: All sentinel errors must include an appropriate exit code from
  `internal/exitcode` following the sysexits.h convention (e.g., `ExUsage` for
  CLI misuse, `ExConfig` for configuration errors, `ExUnavailable` for service
  failures)
- **Testing**: Use `testing` package with table-driven tests; golden files via
  `goldie/v2` for output snapshots. Tests continue to use `errors.Is()` to
  check wrapped errors.
- **Naming**: Unexported types/vars use camelCase, exported use PascalCase;
  struct field tags for JSON (e.g., `json:"model"`)
- **Comments**: All types/funcs (exported and unexported) require doc comments
  starting with the name (e.g., `// LLMClient is an interface...`)
- **Interfaces**: Define minimal interfaces (e.g., `LLMClient` with `Generate`
  method)
- **Context**: Pass `context.Context` as first parameter to functions that make
  external calls
- **Logging**: Use `charmbracelet/log` with appropriate levels:
  - `Error`: Fatal failures, application errors requiring immediate attention
  - `Info`: Important user-facing events (reserved for significant milestones)
  - `Debug`: Internal operations, initialization, config loading, API calls/responses,
    troubleshooting details
- **Configuration**: Ghost uses CLI flags with optional TOML config file support:
  - Config file location: `~/.config/ghost/config.toml`
  - Settings: `host` (Ollama API URL), `model` (LLM model name), `system` (system
   prompt override)
  - Defaults: `host="http://localhost:11434"`, `model="llama3.1:8b"`, `system=""`
   (optional)
  - Use `urfave/cli-altsrc/v3` with `OnlyOnce: true` for flags that load from config
  - Config file is optional; all settings can be provided via CLI flags

## Configuration File Pattern

Ghost loads configuration from `~/.config/ghost/config.toml` if present. The TOML
file structure:

```toml
host = "http://localhost:11434"
model = "llama3.1:8b"
system = "You are Ghost, a cyberpunk inspired terminal based assistant."
```

Flags are defined with the following pattern:

```go
&cli.StringFlag{
    Name:     "host",
    Usage:    "Ollama API URL",
    Value:    "http://localhost:11434",
    Sources:  cli.NewValueSourceChain(toml.TOML("host", configFile)),
    OnlyOnce: true,
}
```

The `OnlyOnce: true` flag ensures the value is loaded only once from the config
file, preventing repeated parsing. Config file loading uses:

```go
homeDir, err := os.UserHomeDir()
configFile := filepath.Join(homeDir, ".config/ghost", "config.toml")
sourcer := altsrc.NewStringPtrSourcer(&configFile)
```

## Error Handling Pattern

### Defining Sentinel Errors

```go
package mypackage

import (
    "errors"
    "github.com/theantichris/ghost/internal/exitcode"
)

var (
    // ErrConfig indicates a configuration error occurred.
    ErrConfig = exitcode.New(errors.New("configuration error"), exitcode.ExConfig)

    // ErrNetwork indicates a network operation failed.
    ErrNetwork = exitcode.New(errors.New("network error"), exitcode.ExUnavailable)
)
```

### Returning Errors

```go
func DoSomething() error {
    if configMissing {
        return fmt.Errorf("%w", ErrConfig)
    }

    if err := networkCall(); err != nil {
        return fmt.Errorf("%w: %w", ErrNetwork, err)
    }

    return nil
}
```

### Extracting Exit Codes (main.go)

```go
if err := cmd.Run(ctx, args, output, logger); err != nil {
    code := exitcode.GetExitCode(err)
    fmt.Printf("error: %s\n", err)
    logger.Error("command failed", "error", err)
    os.Exit(int(code))
}
```

### Testing Errors

```go
func TestSomething(t *testing.T) {
    err := DoSomething()

    // errors.Is() works with wrapped exitcode.Error
    if !errors.Is(err, ErrConfig) {
        t.Errorf("expected ErrConfig, got %v", err)
    }
}
```

## Pre-commit Hooks

Run automatically: `go-fmt`, `go-mod-tidy`, `go-unit-tests`, `golangci-lint`,
`markdownlint`, `codespell`

All pre-commit hooks must pass after editing files.

## Pre-PR Checklist

Before submitting a pull request, ensure all items are completed:

1. **Pre-commit hooks pass**: All hooks run automatically (go-fmt, go-mod-tidy,
 go-unit-tests, golangci-lint, markdownlint, codespell)
2. **Build succeeds**: `go build -v ./...`
3. **Docblocks verified**: All exported and unexported types/functions have
 proper doc comments starting with the name
4. **API documentation**: Verify GoDoc comments render correctly via `go doc`
5. **Error handling correct**: Sentinel errors use `exitcode.New()` with appropriate
 exit codes
6. **Code style followed**: Imports ordered, naming conventions, context passing
7. **Security audit**: No secrets/keys exposed or logged, proper input validation,
 error sanitization
8. **No TODOs remaining**: Search for TODO comments with `rg -i "TODO|FIXME|XXX|HACK"`
 and resolve or document them
9. **README.md updated**: If feature/usage changes, update examples and documentation
10. **AGENTS.md updated**: If new patterns, commands, or guidelines emerge
11. **Golden files updated**: If test output changes, regenerate with
 `go test -update`
12. **Code review**: Review code for best practices, patterns, and potential issues
