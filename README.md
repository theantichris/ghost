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

## Installation

### Via Go Install

```bash
go install github.com/theantichris/ghost@latest
```

### Prebuilt Binaries

Download the latest release for your platform from the
[releases page](https://github.com/theantichris/ghost/releases).

## Usage

### Quick Start

Ghost requires a model to be specified. Run with the `--model` flag:

```bash
ghost --model llama3 "your prompt here"
```

### Basic Examples

```bash
# Get intel on tech
ghost -m llama3 "Explain how neural interfaces work"

# Quick data lookup
ghost -m llama3 "What's the difference between a netrunner and a decker?"

# Code assistance for your next run
ghost -m llama3 "Write a Go function to encrypt data with AES-256"

# Pipe data for analysis
cat error.log | ghost -m llama3 "what's wrong here"
echo "def foo():\n  return bar" | ghost -m llama3 "explain this code"
```

### Help

```bash
ghost --help
```

## Configuration

Ghost can be configured in three ways (in order of precedence):

1. **Command-line flags**
2. **Environment variables**
3. **Configuration file**

### Flags

- `--model`, `-m`: The Ollama model to use (required)
- `--url`, `-u`: Ollama API URL (default: `http://localhost:11434/api`)
- `--config`, `-c`: Path to config file (default: `~/.config/ghost/config.toml`)

### Environment Variables

```bash
export GHOST_MODEL=llama3
export GHOST_URL=http://localhost:11434/api
ghost "your prompt here"
```

### Configuration File

Create `~/.config/ghost/config.toml`:

```toml
model = "llama3"
url = "http://localhost:11434/api"
```

## License

MIT
