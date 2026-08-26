# Ghost

Ghost is a local-first terminal AI companion powered by Ollama.

The project is being rebuilt from first principles with a TUI-first experience and a cyberpunk presentation.

## Prerequisites

- Go 1.26.5
- Ollama installed and running locally
- An Ollama chat model

Ghost currently connects to Ollama at:

`http://localhost:11434/api`

## Model setup

Pull the small model used by the documented examples:

```bash
ollama pull qwen3:0.6b
```

Confirm that it is installed

```bash
ollama ls
```

## One-shot usage

From the repository root:

```bash
go run . --model qwen3:0.6b "Reply with one short sentence confirming Ghost is online."
```

Ghost prints one model-generated response to standard output followed by a newline. The exact response varies between models and invocations.

## Validation

### Normal repository checks

These checks do not require Ollama or network access:

```bash
go build ./...
go test ./...
golangci-lint run
```

### Local end-to-end test

The tagged end-to-end test requires Ollama to be running and the selected model to be installed:

```bash
go test -tags=e2e -v .
```

To test with another installed model:

```bash
GHOST_E2E_MODEL=<model-name> go test -tags=e2e -v .
```

The end-to-end test is intentionally excluded from the normal test suite.

## Troubleshooting

### Ollama is unavailable

Start the Ollama server:

```bash
ollama serve
```

When Ghost cannot connect, the error begins with:

```text
ghost // uplink failure: request Ollama chat response:
send Ollama chat request:
```

### The requested model is missing

List the locally installed models:

```bash
ollama ls
```

Pull the requested model when necessary:

```bash
ollama pull <model-name>
```

A missing model error includes:

```text
ghost // uplink failure: request Ollama chat response:
Ollama chat request failed: 404 Not Found:
```

## Previous version

The previous v3 implementation is preserved on the `archive/v3` branch at the `archive-v3-2026-08-22` tag.
