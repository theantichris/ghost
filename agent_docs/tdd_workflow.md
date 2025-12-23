# TDD Workflow – Ghost

This document describes the Test-Driven Development workflow for Ghost. Use
this as your default process for all coding tasks.

## Red → Green → Refactor Cycle

### 1. Red: Write a Failing Test

**Steps**:

1. Identify the smallest testable behavior you want to add or change
2. Write a test that exercises that behavior
3. Run `go test ./...` to verify the test fails
4. Confirm the failure message is what you expect

**For Ghost**:

- Use table-driven tests with `tests := []struct{...}`
- For output validation, use `goldie/v2` golden files
- Mock LLM calls with `llm.MockLLMClient`
- Use `io.Discard` for loggers in tests

**Example structure**:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  error
    }{
        {name: "case description", input: "test", expected: "result"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 2. Green: Make the Test Pass

**Steps**:

1. Write the **minimum** code to make the test pass
2. Run `go test ./...` to verify
3. Don't add extra features or "nice-to-haves"
4. Focus on making this specific test pass

**For Ghost**:

- Keep `internal/cmd/` focused on CLI logic
- Keep `internal/llm/` focused on LLM operations
- Keep `internal/exitcode/` focused on exit code handling
- Use proper error wrapping with `fmt.Errorf("%w: %w", ...)`
- Return sentinel errors wrapped with exit codes

### 3. Refactor: Improve Without Changing Behavior

**Steps**:

1. Look for duplication or unclear code
2. Improve structure while keeping tests green
3. Run `go test ./...` after each refactor
4. Run `golangci-lint run` to check for issues

**For Ghost**:

- Extract common patterns into helper functions
- Ensure doc comments start with the name
- Keep error chains intact for `errors.Is/As`
- Maintain separation between CLI, LLM, and exit code layers

## Working with Golden Files

### Initial Creation

1. Write test with `goldie.New(t)` and `g.Assert(t, "test_name", output)`
2. Run test (will fail - no golden file)
3. Review actual output carefully
4. Run `go test -update` to create golden file
5. Verify golden file content in `internal/cmd/testdata/`
6. Commit golden file with code

### Updating Existing

1. Make code change
2. Run test (will fail - output changed)
3. Review diff carefully
4. If output is correct, run `go test -update`
5. Verify updated golden file
6. Commit golden file with code

### Common Pitfalls

- Whitespace matters: extra newlines will fail tests
- Golden files are byte-for-byte comparisons
- Review diffs before updating to catch regressions

## Mocking the LLM Client

**Pattern**:

```go
llmClient := llm.MockLLMClient{
    GenerateFunc: func(ctx context.Context, systemPrompt,
        userPrompt string, images []string) (string, error) {
        return "mocked response", nil
    },
    VersionFunc: func(ctx context.Context) (string, error) {
        return "0.1.0", nil
    },
    ShowFunc: func(ctx context.Context, model string) error {
        if model == "missing:model" {
            return llm.ErrModelNotFound
        }
        return nil
    },
}
```

**Usage**:

- Set `GenerateFunc` to control LLM responses
- Set `Error` field to simulate failures for all methods
- Set individual funcs to customize per-method behavior

## Testing Error Handling

**Pattern**:

```go
tests := []struct {
    name    string
    wantErr error
}{
    {name: "returns config error", wantErr: cmd.ErrConfigFile},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        err := functionUnderTest()
        if !errors.Is(err, tt.wantErr) {
            t.Errorf("expected %v, got %v", tt.wantErr, err)
        }
    })
}
```

**Important**:

- Use `errors.Is(err, expected)` for wrapped errors
- Don't match error strings
- Test that proper exit codes are set

## Tutorial Mode Step Template

When guiding a user through TDD:

1. **State the goal**: "We're adding [feature] to [component]"
2. **Propose test**: Show full test code
3. **Instruct**: "Add this test to `file_test.go` and run `go test ./...` to
   verify it fails"
4. **Wait for confirmation**
5. **Propose implementation**: Show full implementation code
6. **Instruct**: "Add this to `file.go` and run `go test ./...` to verify it
   passes"
7. **Wait for confirmation**
8. **Refactor if needed**: Propose improvements
9. **End with**: "Next step: [description]"

## Common Test Patterns in Ghost

### Testing CLI Commands

1. Create mock LLM client
2. Create buffer for output
3. Create test command with metadata
4. Call action function
5. Assert output or error

### Testing LLM Client

1. Use httptest.Server for fake Ollama API
2. Return canned responses
3. Verify requests are correct
4. Test error conditions

### Testing Exit Codes

1. Call function that returns error
2. Use `exitcode.GetExitCode(err)` to extract code
3. Assert code matches expected
4. Verify error chain with `errors.Is()`
