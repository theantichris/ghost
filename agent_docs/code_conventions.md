# Code Conventions – Ghost

## Go Style

### Comments

- All exported types, functions, and methods require doc comments
- Doc comments must start with the name being documented
- Package comments go at the top of the first file in the package

**Example**:

```go
// Ghost represents the main application state.
type Ghost struct { ... }

// Run executes the ghost command with the given context and arguments.
func Run(ctx context.Context, args []string) error { ... }
```

### Context

- Pass `context.Context` as the first parameter for external calls
- Use `ctx` as the parameter name
- Don't store contexts in structs

### Error Handling

**Wrapping**:

- Always use `fmt.Errorf("%w: %w", sentinel, underlying)` for proper error chains
- Define sentinel errors at package level
- Wrap sentinel errors with `exitcode.New()` to set exit codes

**Checking**:

- Use `errors.Is(err, expected)` for sentinel error checking
- Use `errors.As(err, &target)` for type assertions
- Never match error strings

**Pattern**:

```go
// In errors.go
var ErrNoPrompt = exitcode.New(errors.New("prompt is required"), exitcode.ExUsage)

// In implementation
if prompt == "" {
    return fmt.Errorf("%w", ErrNoPrompt)
}

// In tests
if !errors.Is(err, ErrNoPrompt) {
    t.Errorf("expected ErrNoPrompt, got %v", err)
}
```

### Logging

Use `charmbracelet/log` with structured fields:

- `logger.Error("message", "key", value)` - For failures
- `logger.Debug("message", "key", value)` - For internal details
- `logger.Info("message", "key", value)` - Sparingly, for important events

**Never log**:

- Secrets or API keys
- Sensitive user data
- Full prompts (unless debugging with explicit user consent)

### Naming

- Use descriptive names, not single letters (except standard Go idioms: `i`,
  `err`, `ctx`)
- Acronyms should be uppercase: `LLMClient`, `URL`, `HTTP`
- Unexported fields/functions use camelCase: `generateURL`, `defaultModel`
- Exported types/functions use PascalCase: `Generate`, `Ollama`

### Testing

**Table-driven tests**:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  error
    }{
        {name: "description", input: "test", expected: "result"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

**Golden files**:

- Use `goldie/v2` for snapshot testing
- Store in `internal/cmd/testdata/TestName/subtest_name.golden`
- Update with `go test -update`

**Mocking**:

- Use `llm.MockLLMClient` for LLM operations
- Use `io.Discard` for loggers in tests

## Project-Specific Patterns

### CLI Framework (urfave/cli/v3)

**Metadata pattern**:

```go
// Store shared state in metadata
cmd.Metadata["logger"] = logger
cmd.Metadata["llmClient"] = llmClient

// Access in actions
logger := cmd.Root().Metadata["logger"].(*log.Logger)
```

**Flag sources**:

```go
&cli.StringFlag{
    Name:     "host",
    Value:    "http://localhost:11434",
    Sources:  cli.NewValueSourceChain(toml.TOML("host", configFile)),
    OnlyOnce: true,
}
```

### Exit Codes

Always use `internal/exitcode` for exit status:

- `ExUsage` (64): CLI misuse, bad arguments
- `ExDataErr` (65): Invalid input data
- `ExConfig` (78): Configuration error
- `ExUnavailable` (69): Service unavailable (Ollama not running)
- `ExSoftware` (70): Internal software error

**Define at package level**:

```go
var ErrConfig = exitcode.New(errors.New("configuration error"), exitcode.ExConfig)
```

**Extract at boundary** (main.go):

```go
if err := cmd.Run(ctx, args, output, logger); err != nil {
    code := exitcode.GetExitCode(err)
    os.Exit(int(code))
}
```

### Dual Model System

Ghost uses two models - respect this separation:

1. **Vision model**: Analyzes images, returns text description
2. **Chat model**: Processes text prompts (including vision output)

**Flow**:

- If images provided → call vision model first
- Append vision response to user prompt
- Call chat model with enriched prompt

**Don't**:

- Assume a single model can do both
- Skip the vision step if images are provided
- Try to merge the models into one call

### Configuration Loading

Config file at `~/.config/ghost/config.toml` is **optional**:

- If missing, use flag defaults
- `loadConfigFile()` returns valid `StringSourcer` either way
- Don't error on missing config

**Priority**: CLI flags > TOML values > defaults

### Logger Initialization

Logger is initialized in `main.go` and passed via metadata:

- JSON format for machine readability
- File output to `~/.config/ghost/ghost.log`
- `ReportCaller: true` for debugging
- `Level: DebugLevel` to capture all details

## Formatting and Linting

### Pre-commit Hooks

Install with `pre-commit install`. Hooks run:

- `go-fmt`: Format code
- `go-mod-tidy`: Clean dependencies
- `go-unit-tests`: Run test suite
- `golangci-lint`: Lint code
- `markdownlint`: Lint markdown files
- `codespell`: Spell check
- `trailing-whitespace`: Remove trailing whitespace
- `check-yaml`: Validate YAML files

### golangci-lint

Config in `.golangci.yml`:

- Excludes `fmt.Fprintf/Fprintln/Fprint` from errcheck (intentional)
- Run with `golangci-lint run`

### Spell Checking

Custom dictionary in `.harper-dictionary.txt`:

- Add project-specific terms
- Run via pre-commit or `codespell`

## File Organization

### Package Structure

```text
internal/
├── cmd/          # CLI command logic
│   ├── main.go       # Root command, flags, actions
│   ├── health.go     # Health check subcommand
│   ├── config.go     # Config loading
│   ├── errors.go     # Command errors
│   ├── main_test.go  # Tests
│   └── testdata/     # Golden files
├── llm/          # LLM client abstraction
│   ├── client.go     # Interface
│   ├── ollama.go     # Implementation
│   ├── mock.go       # Test mock
│   ├── errors.go     # LLM errors
│   └── *_test.go     # Tests
└── exitcode/     # Exit code handling
    ├── exitcode.go       # Types and functions
    └── exitcode_test.go  # Tests
```

### Separation of Concerns

- `internal/cmd/`: CLI orchestration only, no business logic
- `internal/llm/`: LLM operations, no CLI knowledge
- `internal/exitcode/`: Exit code handling, no domain knowledge

Don't mix concerns across package boundaries.
