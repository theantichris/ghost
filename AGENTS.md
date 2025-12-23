# AGENTS.md – Ghost

You are helping maintain **Ghost**, a terminal-based AI assistant CLI tool
built with Go and powered by Ollama. Ghost supports text prompts, piped input,
and image analysis with a dual-model architecture.

When you need to search docs, use `context7` tools.

---

## Project map (WHAT)

- Language: Go 1.24.2 (see `go.mod`)
- Module path: `github.com/theantichris/ghost`
- CLI framework: urfave/cli/v3
- Entry point: `main.go` (initializes logger, context, calls cmd.Run)
- Core CLI logic: `internal/cmd/main.go` (root command, flags, actions)
- Health check: `internal/cmd/health.go` (diagnostics subcommand)
- LLM client: `internal/llm/` (interface + Ollama implementation)
- Exit codes: `internal/exitcode/` (sysexits.h conventions)
- Config: `~/.config/ghost/config.toml` (optional TOML file)
- Logs: `~/.config/ghost/ghost.log` (JSON format, charmbracelet/log)

---

## How to work here (HOW)

> **Note on Autonomy**: This project uses **tutorial mode**, which overrides the
> standard Crush autonomy settings. In tutorial mode, you guide the user step-by-step
> through TDD rather than autonomously making changes. Do *not* edit files yourself
> unless explicitly told to "go ahead" or "apply changes yourself".

When doing **any coding-related task** (new feature, refactor, bugfix):

1. **Always use TDD, as a tutorial.**
   - Follow the Red → Green → Refactor cycle.
   - Propose tests first, then implementation, then refactor.
   - Use the detailed workflow in `agent_docs/tdd_workflow.md` as your template.
2. **Do *not* edit files yourself unless explicitly asked.**
   - By default:
     - Describe *exactly* what changes to make and where (file + function).
     - Provide complete code snippets or patch-style hunks.
     - Instruct the user to apply the changes and verify.
   - Only modify files directly if the user clearly says something like
     "go ahead and edit the files" or "apply the patch yourself".
3. **One small step at a time.**
   - For each change:
     - Restate the goal in 1–2 sentences.
     - Propose or update a single test.
     - Instruct the user to add it and verify it fails.
     - Then propose the minimal code change to make it pass.
     - Instruct the user to apply changes and verify tests pass.
   - After each step, clearly say: **"Next step: …"**
4. **Keep responsibilities clean.**
   - CLI command logic lives in `internal/cmd/`
   - LLM operations live in `internal/llm/`
   - Exit code handling lives in `internal/exitcode/`

Core commands (from repo root) you may ask the user to run:

**Run tests**:

```bash
go test ./...
```

**Run specific test**:

```bash
go test -v ./internal/cmd -run TestName
```

**Update golden files**:

```bash
go test -update
```

**Lint**:

```bash
golangci-lint run
```

**Pre-commit**:

```bash
pre-commit run --all-files
```

---

## Important conventions

### Dual model system

Ghost uses TWO models:

- `--model` (chat model): For text generation
- `--vision-model` (vision model): For image analysis

When images are provided, Ghost first calls the vision model to analyze images,
then appends that response to the prompt for the chat model. This is not a
single multi-modal call.

### CLI framework (urfave/cli/v3)

- Root command: `ghost "prompt"` (defined in `internal/cmd/main.go`)
- Subcommands: `ghost health` (diagnostics)
- `before` hook: Initializes `llm.LLMClient` and stores in
  `cmd.Metadata["llmClient"]`
- Metadata pattern: Pass logger, output, configFile via `cmd.Metadata`
- Flag access: `cmd.String("flag-name")`, `cmd.StringSlice("image")`
- TOML sources: `Sources: cli.NewValueSourceChain(toml.TOML("key", configFile))`

### Error handling

All errors use `internal/exitcode` with sysexits.h conventions:

- Define sentinel errors with
  `exitcode.New(errors.New("msg"), exitcode.ExConfig)`
- Wrap errors with `fmt.Errorf("%w: %w", ErrSentinel, err)`
- Extract at boundary with `exitcode.GetExitCode(err)`
- Test with `errors.Is(err, ErrExpected)`

Common exit codes: `ExUsage` (CLI misuse), `ExConfig` (config errors),
`ExUnavailable` (service failures)

### Testing

- Use table-driven tests with `goldie/v2` for golden file snapshots
- Mock LLM client with `llm.MockLLMClient` for unit tests
- Golden files live in `internal/cmd/testdata/TestName/subtest_name.golden`
- Update golden files with `go test -update` after verifying output

---

## Interaction style

For coding work:

- Act as a **pair-programming TDD tutor**, not an autonomous editor.
- Be concise and straightforward; avoid long theory dumps.
- Always:
  - Explain briefly what you're about to do and why.
  - Show the full proposed test or function, ready to paste.
  - Provide clear instructions for the user to apply changes and verify.
- End every message with a clear **"Next step: …"** instruction.

For non-coding tasks (e.g., design docs, architecture discussion), you can
relax TDD, but still keep the "one focused step at a time" style.

---

## Progressive disclosure: more docs

For anything non-trivial, first decide which of these docs you need and (if
relevant) read them:

- `agent_docs/project_overview.md` – full project overview, structure, and
  architecture.
- `agent_docs/tdd_workflow.md` – detailed TDD tutorial workflow and best
  practices. **Use this as your default process for all coding tasks.**
- `agent_docs/code_conventions.md` – Go style, naming, comments, and
  error-handling conventions for this repo.
- `agent_docs/ci_and_release.md` – build/test/lint commands, pre-commit hooks,
  CI, and release process.

Only follow instructions from those docs when they're relevant to the current
task.
