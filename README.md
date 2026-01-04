# Ghost

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

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/ghost.svg)](https://pkg.go.dev/github.com/theantichris/ghost)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/markdown.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost)
![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

Ghost is a local, AI assistant tool built with Go and powered by Ollama.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI
companion into a terminal-first experience.

## Prerequisites

- [Ollama](https://ollama.ai) installed and running
- A model of your choice pulled
- (Optional) A vision model for image analysis (e.g., `llama3.2-vision`, `llava`)

## Installation

### Via Go Install

```bash
go install github.com/theantichris/ghost@latest
```

### Prebuilt Binaries

Download the latest release for your platform from the
[releases page](https://github.com/theantichris/ghost/releases).

## Usage

```bash
# Get intel on tech
ghost "Explain how neural interfaces work"

# Quick data lookup
ghost "What's the difference between a netrunner and a decker?"

# Code assistance for your next run
ghost "Write a Go function to encrypt data with AES-256"

# Get structured JSON output with syntax highlighting
ghost -f json "List the top 3 programming languages"

# JSON output can be piped to other tools (no color codes)
ghost -f json "system info" | jq .

# Get formatted markdown output with cyberpunk theme
ghost -f markdown "Write a guide to memory management"

# Markdown output can be piped to other tools (no color codes)
ghost -f markdown "Write a guide to memory management" >> memory.md

# Pipe data for analysis
cat error.log | ghost "what's wrong here"
echo "def foo():\n  return bar" | ghost "explain this code"

# Analyze images (requires vision model)
ghost -i screenshot.png "what's in this image?"

# Analyze multiple images
ghost -i img1.png -i img2.png "compare these images"

# Use specific vision model
ghost -V llama3.2-vision -i diagram.png "explain this diagram"
```

## Configuration

Ghost can be configured in three ways (in order of precedence):

1. **Command-line flags**
2. **Environment variables**
3. **Configuration file**

### Flags

- `--model`, `-m`: The Ollama model to use
- `--vision-model`, `-V`: Vision model for image analysis (defaults to main model)
- `--image`, `-i`: Path to image file(s) for analysis (can be specified multiple
times)
- `--url`, `-u`: Ollama API URL (default: `http://localhost:11434/api`)
- `--format`, `-f`: Output format (default: text, options: json, markdown)
- `--config`, `-c`: Path to config file (default: `~/.config/ghost/config.toml`)

### Environment Variables

```bash
export GHOST_MODEL=llama3
export GHOST_VISION_MODEL=llama3.2-vision
export GHOST_URL=http://localhost:11434/api

ghost "your prompt here"
```

### Configuration File

Create `~/.config/ghost/config.toml`:

```toml
model = "llama3"
vision_model = "llama3.2-vision"
url = "http://localhost:11434/api"
```

## License

MIT
