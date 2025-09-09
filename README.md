# assistant-go

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/assistant-go.svg)](https://pkg.go.dev/github.com/theantichris/assistant-go) [![CI](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml/badge.svg)](https://github.com/theantichris/assistant-go/actions/workflows/ci.yml)

assistant-go is a local, general-purpose AI assistant and orchestrator written in Go and powered by a locally running Ollama instance. It is designed for research, chat, and task automation with optional, explicit opt-in connectivity. It has a UI/UX based on classic cyberpunk media like Shadowrun, Cyberpunk 2077, and the Matrix.

## Documents

- Specification: [SPEC.md](SPEC.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)
- LLM integration details: [LLMS.md](LLMS.md)
- Agents, tools, and memory: [AGENTS.md](AGENTS.md)

## Requirements

- Go >= 1.24
- Ollama installed and running locally
- At least one model pulled (e.g., `ollama pull llama3.1`)

## Environment variables

- OLLAMA_MODEL
- OLLAMA_HOST

## Examples

```go
go run ./cmd/assistant -model llama3.1 "Hello"
echo "Hello" | go run ./cmd/assistant -model llama3.1
export OLLAMA_MODEL=llama3.1; export OLLAMA_HOST=http://localhost:11434; go run ./cmd/assistant "Hello from env"
```
