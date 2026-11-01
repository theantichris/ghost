# Repository Guidelines

## Project Structure & Module Organization

Ghost is a local-first terminal AI companion using Ollama. The current v4 implementation supports one-shot prompts; the README describes the planned TUI direction.

- `main.go` handles CLI arguments, session creation, output, and errors.
- `internal/conversation/` owns messages, session history, and turn execution.
- `internal/ollama/` implements the HTTP chat client.
- Tests live beside implementation files as `*_test.go`.
- `.github/workflows/go.yml` builds and tests pull requests targeting `main`.

There is no separate asset directory. Keep new functionality within the relevant package rather than expanding CLI responsibilities.

## Build, Test, and Development Commands

Use the Go version declared in `go.mod` (currently 1.26.5).

- `go build ./...`: compile every package.
- `go test ./...`: run offline tests without a live Ollama instance.
- `golangci-lint run`: run Go static analysis.
- `go run . --model qwen3:0.6b "Hello"`: send a prompt to local Ollama at `http://localhost:11434/api`.
- `go test -tags=e2e -v .`: run the live workflow test with an installed model. Set `GHOST_E2E_MODEL` to select another model.

The configured pre-push hook also runs the live test; CI runs only build and offline tests.

## Coding Style & Naming Conventions

Use `gofmt` for Go formatting, including tab indentation. Keep package names lowercase and exported identifiers in Go-style PascalCase. Wrap errors with operation context using `%w`, following existing code.

The pre-commit configuration includes Go checks, Markdown linting, spelling, and whitespace checks. Markdown uses standard headings; the line-length rule is disabled.

## Testing Guidelines

Use Go's standard `testing` package, `TestXxx` names, and descriptive table-driven subtests. Inject chat functions and use `httptest` for isolated HTTP behavior. Cover changed behavior and failure paths; no numeric coverage threshold is configured. Keep live inference tests behind the `e2e` build tag.

## Commit & Pull Request Guidelines

Recent history mixes `chore:` and `ci:` prefixes with issue-linked feature subjects such as `Feat/issue 298/preserve context`. Use a concise, descriptive subject and reference the relevant issue. Pull requests should explain behavior changes and validation results; include terminal output or screenshots when presentation changes.

## Agent Working Agreements

Christopher writes the project code. Agent acts as a tutor and reviewer to help Christopher refresh coding skills and learn Mitos. Work one small step at a time: explain the goal, let Christopher implement it, then review and validate together. Offer hints before full solutions; provide full solutions when Christopher asks. Include relevant Mitos navigation, editing, and diagnostics guidance during development.

Before starting an issue, Agent reviews its scope, prerequisites, and sequencing against the current code.

Agent gives Christopher one concrete implementation step at a time and explains its purpose. When Christopher says "done," Agent reads the saved changes, gives specific feedback, and resolves any mistakes with Christopher before introducing the next step.

Christopher writes production code. Agent may write and maintain tests, following the repository's existing conventions and style.

Christopher prefers small, reversible commits. Agent keeps implementation, tests, and documentation as separate reviewable checkpoints, pauses between them so Christopher can commit, and suggests concise commit messages.

Agent runs validation at meaningful checkpoints rather than after every small edit. Agent distinguishes completed issue scope from functionality intentionally scheduled for later issues.

Do not make implementation edits unless Christopher explicitly requests them. When reviewing, explain concrete findings and let Christopher apply the fixes.

Agent should explain substantial changes to Christopher, preserve unrelated work, and report checks that could not run. Do not commit, push, or install packages unless Christopher explicitly requests it.
