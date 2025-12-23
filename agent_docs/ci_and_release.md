# CI and Release – Ghost

## Local Development Commands

### Build

```bash
go build -v ./...
```

Builds all packages with verbose output.

### Test

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v ./internal/cmd -run TestName

# Run specific subtest
go test -v ./internal/cmd -run TestName/subtest_name

# Run E2E tests (requires ollama in PATH)
go test -tags=e2e -v .

# Update golden files
go test -update
```

### Lint

```bash
golangci-lint run
```

Runs all configured linters. Config in `.golangci.yml`.

### Format

```bash
go fmt ./...
```

Formats all Go files. Also runs automatically in pre-commit.

### Pre-commit Hooks

```bash
# Install hooks
pre-commit install

# Run all hooks manually
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint
```

Hooks configured in `.pre-commit-config.yaml`:

- `go-fmt`: Format Go code
- `go-mod-tidy`: Clean up go.mod/go.sum
- `go-unit-tests`: Run test suite
- `golangci-lint`: Lint code
- `markdownlint`: Lint markdown files (with `--fix`)
- `codespell`: Spell check (uses `.harper-dictionary.txt`)
- `trailing-whitespace`: Remove trailing whitespace (excludes `.golden`)
- `end-of-file-fixer`: Ensure newline at EOF (excludes `.golden`)
- `check-yaml`: Validate YAML syntax
- `check-added-large-files`: Block files >5MB
- `check-merge-conflict`: Detect merge conflict markers

### Find TODOs

```bash
rg -i "TODO|FIXME|XXX|HACK"
```

Searches for common task markers in code.

## GitHub Actions Workflows

### go.yml

Runs on push/PR to `main`:

1. Checkout code
2. Set up Go 1.24
3. Build with `go build -v ./...`
4. Test with `go test -v ./...`

### markdown.yml

Lints markdown files on push/PR to `main`.

### release.yml

Triggered by pushing a git tag:

1. Runs GoReleaser
2. Builds binaries for Linux, macOS, Windows
3. Creates GitHub release with artifacts
4. Archives as `.tar.gz` (Unix) or `.zip` (Windows)

## Release Process

### Version Management

Version is injected at build time via `-ldflags`:

```go
// In main.go
var version = "dev"
```

GoReleaser sets this with `-ldflags -X main.version={{.Version}}`.

For local builds, version defaults to "dev".

### Creating a Release

1. Ensure all tests pass: `go test ./...`
2. Ensure lint passes: `golangci-lint run`
3. Update CHANGELOG (if applicable)
4. Create and push tag:

   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

5. GitHub Actions runs `release.yml` workflow
6. GoReleaser builds and publishes artifacts

### GoReleaser Configuration

Configured in `.goreleaser.yaml`:

**Before hooks**:

- `go mod tidy`

**Build settings**:

- `CGO_ENABLED=0` for static binaries
- Targets: Linux, macOS, Windows
- LDFLAGS: `-s -w -X main.version={{.Version}}`

**Archives**:

- Format: `.tar.gz` for Unix, `.zip` for Windows
- Name template: `ghost_<OS>_<Arch>`

**Changelog**:

- Auto-generated
- Excludes commits starting with `docs:` or `test:`

### Manual Release Testing

Test release locally without publishing:

```bash
goreleaser release --snapshot --clean
```

This builds for all platforms in `./dist/` without creating a release.

## E2E Tests

E2E tests in `e2e_test.go`:

- Build a temporary binary with `go build`
- Execute the binary with test cases
- Validate stdout, stderr, and exit codes
- Require `ollama` in PATH (skip if not found)
- Clean up binary after tests

Run manually:

```bash
go test -v -run TestE2E
```

## Spell Checking

Custom dictionary in `.harper-dictionary.txt` for project-specific terms:

- `Ollama`
- `llama`
- `qwen`
- `cyberpunk`
- etc.

Add new terms as needed when codespell flags them.

## Dependencies

Key dependencies (from `go.mod`):

- `github.com/urfave/cli/v3` - CLI framework
- `github.com/urfave/cli-altsrc/v3` - TOML config loading
- `github.com/charmbracelet/log` - Structured logging
- `github.com/carlmjohnson/requests` - HTTP client
- `github.com/sebdah/goldie/v2` - Golden file testing
- `github.com/BurntSushi/toml` - TOML parsing

Update dependencies:

```bash
go get -u ./...
go mod tidy
```

## Troubleshooting

### Tests failing locally but passing in CI

- Check Go version: `go version` (should be 1.24.2)
- Ensure pre-commit hooks pass: `pre-commit run --all-files`
- Check for uncommitted golden file changes

### Golden file mismatches

- Review actual vs expected output
- If correct, update: `go test -update`
- Commit updated golden files

### Lint failures

- Run `golangci-lint run` locally
- Fix reported issues
- Some checks excluded in `.golangci.yml` (e.g., `fmt.Fprintf` errcheck)

### E2E tests skipped

- E2E tests require `ollama` in PATH
- Install Ollama: <https://ollama.ai>
- Verify with `ollama --version`
