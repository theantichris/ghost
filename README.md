# Ghost

  ▄████  ██░ ██  ▒█████   ██████  ████████
  ██▒ ▀█▒▓██░ ██▒▒██▒  ██▒ ██    ▒    ██
 ▒██░▄▄▄░▒██▀▀██░▒██░  ██▒ ▓██▄       ██
 ░▓█  ██▓░▓█ ░██ ▒██   ██░  ▒   ██▒   ██
 ░▒▓███▀▒░▓█▒░██▓░ ████▓▒░▒██████▒▒   ██
  ░▒   ▒   ▒ ░░▒░▒░ ▒░▒░▒░ ▒ ▒▓▒ ▒ ░   ░
   ░   ░   ▒ ░▒░ ░  ░ ▒ ▒░ ░ ░▒  ░ ░
 ░ ░   ░   ░  ░░ ░░ ░ ░ ▒  ░  ░  ░
       ░   ░  ░  ░    ░ ░        ░

> **One prompt enters the wire. One local response comes back.**

Ghost is a local-first terminal AI companion powered by Ollama.

The project is being rebuilt from first principals as a TUI-first cyberdeck: local inference, precise terminal behavior, and a cyberpunk interface designed for the glow of the command line.

## Boot Sequence

Before opening the uplink, install:

- Go 1.26.5
- Ollama, running locally
- An Ollama chat model

Ghost currently expects the Ollama API at: `http://localhost:11434/api`

## Load a Construct

Pull the small model used by the documented examples:

```bash
ollama pull qwen3:0.6b
```

Confirm that it is available in local model storage:

```bash
ollama ls
```

## Open the uplink

From the repository root:

```bash
go run . --model qwen3:0.6b "Reply with one short sentence confirming Ghost is online."
```

On a successful connection, Ghost prints one model-generated response to standard output followed by a newline.

The exact response varies between models and invocations. The signal is nondeterministic; the pathway is not.

## System Diagnostics

Ghost has two validation circuits: isolated repository checks and an explicit live connection test.

### Offline checks

These commands do not require Ollama or network access:

```bash
go build ./...
go test ./...
golangci-lint run
```

### Live uplink check

The tagged end-to-end test connects to a real local Ollama instance. Ollama must be running, and the selected model must already be installed

```bash
go test -tags=e2e -v .
```

To validate another installed model:

```bash
GHOST_E2E_MODEL=<model-name> go test -tags=e2e -v .
```

The end-to-end test is intentionally excluded from the normal test suite and continuous integration. Nothing reaches the live wire unless explicitly requested.

## Signal Recovery

### Ollama node unavailable

Start the Ollama server:

```bash
ollama serve
```

When Ghost cannot establish a connection, the error chain begins with:

```text
ghost // uplink failure: request Ollama chat response:
send Ollama chat request:
```

The remaining text comes from the operating system and may differ between platforms.

### Requested model missing

Inspect the models available in local storage:

```bash
ollama ls
```

Pull the requested model when necessary:

```bash
ollama pull <model-name>
```

A missing-model failure includes:

```text
ghost // uplink failure: request Ollama chat response:
Ollama chat request failed: 404 Not Found:
```

## Archived Construct

The previous v3 implementation is preserved on the `archive/v3` branch and at the `archive-v3-2026-08-22` tag.
