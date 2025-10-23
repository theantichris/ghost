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

[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/ghost.svg)](https://pkg.go.dev/github.com/theantichris/ghost)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Build Status](https://github.com/theantichris/ghost/actions/workflows/markdown.yml/badge.svg)](https://github.com/theantichris/ghost/actions)
[![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost)
![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

Ghost is a local, AI assistant CLI tool built-in Go and powered by Ollama.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI
companion into a terminal-first experience.

## Prerequisites

- [Ollama](https://ollama.ai) installed and running
- A model of your choice pulled (defaults to `llama3.1:8b`)

## Installation

### Via Go Install

```bash
go install github.com/theantichris/ghost@latest
```

### Pre-built Binaries

Download the latest release for your platform from the
[releases page](https://github.com/theantichris/ghost/releases).

## Usage

### Quick Start

Invoke Ghost with a prompt to get instant AI assistance in your terminal:

```bash
ghost "your prompt here"
```

### Basic Examples

```bash
# Get intel on tech
ghost "Explain how neural interfaces work"

# Quick data lookup
ghost "What's the difference between a netrunner and a decker?"

# Code assistance for your next run
ghost "Write a Go function to encrypt data with AES-256"
```

### Health Check

Run diagnostics to verify Ghost is properly configured and connected:

```bash
ghost health
```

The health command performs comprehensive system checks:

- **System Configuration**: Displays active host, model, and config file location
- **Neural Link Status**: Verifies Ollama API connectivity and version
- **Model Validation**: Confirms the configured model is loaded and available

#### Example Output (All Systems Nominal)

```text
>> initializing ghost diagnostics...

SYSTEM CONFIG
  ◆ host: http://localhost:11434
  ◆ model: llama3.1:8b
  ◆ config: /home/user/.config/ghost/config.toml

NEURAL LINK STATUS
  ◆ ollama api CONNECTED [v0.1.32]
  ◆ model llama3.1:8b active

>> ghost online :: all systems nominal
```

#### Example Output (Critical Errors)

```text
>> initializing ghost diagnostics...

SYSTEM CONFIG
  ◆ host: http://localhost:11434
  ◆ model: llama3.1:8b
  ◆ config:

NEURAL LINK STATUS
  ✗ ollama api CONNECTION FAILED: connection refused
  ✗ model llama3.1:8b not loaded: connection refused

>> ghost offline :: 2 critical errors detected
```

### Using Custom Configuration

```bash
# Use a specific model
ghost --model "codellama:13b" "write a function to parse JSON"

# Connect to a remote Ollama instance
ghost --host "http://192.168.1.50:11434" "your prompt here"

# Override system prompt for specialized tasks
ghost --system "You are an expert Go developer" "how do I handle context cancellation?"

# Check health with custom configuration
ghost --host "http://192.168.1.50:11434" --model "codellama:13b" health
```

### Help

```bash
ghost --help
ghost health --help
```

## Configuration

Ghost can be configured via CLI flags or an optional TOML configuration file.

### CLI Flags

- `--host`: Ollama API URL (default: `http://localhost:11434`)
- `--model`: LLM model name (default: `llama3.1:8b`)
- `--system`: System prompt override (optional)

### Configuration File

Create a config file at `~/.config/ghost/config.toml`:

```toml
host = "http://localhost:11434"
model = "llama3.1:8b"
system = "You are Ghost, a cyberpunk inspired terminal based assistant."
```

Settings in the config file are used as defaults. CLI flags override config file
values.

### Examples with Flags

```bash
# Use a different Ollama host
ghost --host "http://192.168.1.100:11434" "your prompt"

# Use a different model
ghost --model "dolphin-mixtral:8x7b" "your prompt"

# Override the system prompt
ghost --system "You are a helpful coding assistant" "explain async/await in Go"

# Combine multiple flags
ghost --host "http://remote:11434" --model "llama3.1:70b" "your prompt"
```

## License

MIT
