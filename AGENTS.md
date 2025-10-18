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
- **Errors**: Define sentinel errors as package-level vars (e.g.,
  `var ErrOutput = errors.New("...")`), wrap with
  `fmt.Errorf("%w: %w", ErrParent, err)`
- **Testing**: Use `testing` package with table-driven tests; golden files via
  `goldie/v2` for output snapshots
- **Naming**: Unexported types/vars use camelCase, exported use PascalCase;
  struct field tags for JSON (e.g., `json:"model"`)
- **Comments**: All types/funcs (exported and unexported) require doc comments
  starting with the name (e.g., `// LLMClient is an interface...`)
- **Interfaces**: Define minimal interfaces (e.g., `LLMClient` with `Generate`
  method)
- **Context**: Pass `context.Context` as first parameter to functions that make
  external calls

## Pre-commit Hooks

Run automatically: `go-fmt`, `go-mod-tidy`, `go-unit-tests`, `golangci-lint`,
`markdownlint`, `codespell`

All pre-commit hooks must pass after editing files.
