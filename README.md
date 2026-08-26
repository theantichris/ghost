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

```test
ollam pull qwen3:0.6b
```
