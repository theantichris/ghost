# Agent Guidelines for Ghost

When you need to search docs, use `context7` tools.

## Quick Commands

- **Build**: `go build -v ./...`
- **Test**: `go test -v ./...` (single: `go test -v ./internal/cmd -run TestName`)
- **Lint**: `golangci-lint run`
- **Update golden files**: `go test -update`
- **Find TODOs**: `rg -i "TODO|FIXME|XXX|HACK"`
- **Pre-commit setup**: `pre-commit install`
- **Run pre-commit**: `pre-commit run --all-files`

## Code Style Essentials

- **Go version**: 1.24.2
- **Comments**: All types/funcs require doc comments starting with the name
- **Context**: Pass `context.Context` as first parameter for external calls
- **Logging**: Use `charmbracelet/log` (`Error` for failures, `Debug` for internals,
 `Info` sparingly)
- **Testing**: Table-driven tests with `goldie/v2` golden files
- **Pre-commit**: Uses pre-commit hooks for code quality
- **Spell checking**: Custom dictionary in `.harper-dictionary.txt`
- **Golden files**: Use `goldie/v2` for snapshot testing

## Configuration Pattern

Configuration uses TOML files with the following structure:
- Root level: basic chat settings (host, model, system)
- [vision] section: image analysis settings
- CLI flags override config file values
- Config location: `~/.config/ghost/config.toml`

See [README.md](README.md#configuration) for user-facing details. For code:

## Error Handling Pattern

All errors use `internal/exitcode` with sysexits.h conventions:

```go
// Define sentinel errors with exit codes
var ErrConfig = exitcode.New(errors.New("configuration error"), exitcode.ExConfig)

// Wrap errors when returning
func DoSomething() error {
    if err := operation(); err != nil {
        return fmt.Errorf("%w: %w", ErrConfig, err)
    }
    return nil
}

// Extract exit codes at application boundary (main.go)
if err := cmd.Run(ctx, args, output, logger); err != nil {
    code := exitcode.GetExitCode(err)
    os.Exit(int(code))
}

// Test with errors.Is()
if !errors.Is(err, ErrConfig) {
    t.Errorf("expected ErrConfig, got %v", err)
}
```

Common exit codes: `ExUsage` (CLI misuse), `ExConfig` (config errors), `ExUnavailable`
 (service failures)
