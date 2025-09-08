# assistant-go

A local conversational AI assistant and orchestrator built in Go, powered by Ollama.

Status: MVP CLI named `assistant` that sends a single user message to a locally running Ollama `/api/chat` endpoint and prints the assistant response. Streaming is enabled by default.

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
