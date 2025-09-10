# ghost

```text
   ▄████  ██░ ██  ▒█████   ██████  ████████
  ██▒ ▀█▒▓██░ ██▒▒██▒  ██▒ ██    ▒    ██
 ▒██░▄▄▄░▒██▀▀██░▒██░  ██▒ ▓██▄       ██
 ░▓█  ██▓░▓█ ░██ ▒██   ██░  ▒   ██▒   ██
 ░▒▓███▀▒░▓█▒░██▓░ ████▓▒░▒██████▒▒   ██
  ░▒   ▒   ▒ ░░▒░▒░ ▒░▒░▒░ ▒ ▒▓▒ ▒ ░   ░
   ░   ░   ▒ ░▒░ ░  ░ ▒ ▒░ ░ ░▒  ░ ░
 ░ ░   ░   ░  ░░ ░░ ░ ░ ▒  ░  ░  ░
       ░   ░  ░  ░    ░ ░        ░
```

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/assistant-go.svg)](https://pkg.go.dev/github.com/theantichris/assistant-go) [![CI](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml/badge.svg)](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml)

**Ghost** is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. It is designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_, _Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always on AI companion into a terminal-first experience.

## Documents

- Specification: [SPEC.md](SPEC.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)

## Requirements

- Go >= 1.24
- Ollama installed and running locally
- At least one model pulled (e.g., `ollama pull llama3.1`)

## Environment variables

- OLLAMA_BASE_URL
- DEFAULT_MODEL

## Examples

```go
echo "Hello" | go run ./cmd/assistant -model llama3.1
go run ./cmd/assistant -model llama3.1 "Hello"
echo "Hello" | go run ./cmd/assistant -model llama3.1
export DEFAULT_MODEL=llama3.1; export OLLAMA_BASE_URL=http://localhost:11434; go run ./cmd/assistant "Hello from env"
```
