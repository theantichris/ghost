# assistant-go

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/assistant-go.svg)](https://pkg.go.dev/github.com/theantichris/assistant-go) [![CI](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml/badge.svg)](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml)

This project is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. Designed for research, chat, and task automation, it runs on your own machine with hybrid connectivity. The vision is inspired by cyberpunk media such as Shadowrun, Cyberpunk 2077, and The Matrix — bringing a versatile, always-on AI companion into a terminal-first experience.

Requirements:

- Go >= 1.24
- Ollama installed and running locally (default `http://localhost:11434`)
- At least one model pulled (e.g., `ollama pull llama3.1`)

Build:

- go build -o ./bin/assistant ./cmd/assistant

Usage:

- assistant -model MODEL [flags] [PROMPT...]
- echo "Hello" | assistant -model MODEL

Flags:

- -model string (required unless env `OLLAMA_MODEL` is set)
- -host string (default from env `OLLAMA_HOST` or `http://localhost:11434`)
- -timeout dur (default `2m`)
- -stream bool (default `true`)

Environment variables:

- OLLAMA_MODEL
- OLLAMA_HOST

Examples:

- go run ./cmd/assistant -model llama3.1 "Hello"
- echo "Hello" | go run ./cmd/assistant -model llama3.1
- export OLLAMA_MODEL=llama3.1; export OLLAMA_HOST=<http://localhost:11434>; go run ./cmd/assistant "Hello from env"
