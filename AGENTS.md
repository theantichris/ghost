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

1. **All tests pass**: `go test -v ./...`
2. **Build succeeds**: `go build -v ./...`
3. **Code is formatted**: `go fmt ./...`
4. **Dependencies are tidy**: `go mod tidy`
5. **Linting passes**: `golangci-lint run`
6. **Pre-commit hooks pass**: All hooks (go-fmt, go-mod-tidy, go-unit-tests,
 golangci-lint, markdownlint, codespell)
7. **Docblocks verified**: All exported and unexported types/functions have
 proper doc comments starting with the name
8. **API documentation**: Verify GoDoc comments render correctly via `go doc`
9. **Error handling correct**: Sentinel errors use `exitcode.New()` with appropriate
 exit codes
10. **Code style followed**: Imports ordered, naming conventions, context passing
11. **Security audit**: No secrets/keys exposed or logged, proper input validation,
 error sanitization
12. **No TODOs remaining**: Search for TODO comments with `rg -i "TODO|FIXME|XXX|HACK"`
 and resolve or document them
13. **README.md updated**: If feature/usage changes, update examples and documentation
14. **AGENTS.md updated**: If new patterns, commands, or guidelines emerge
15. **Golden files updated**: If test output changes, regenerate with
 `go test -update`
16. **Code review**: Review code for best practices, patterns, and potential issues
