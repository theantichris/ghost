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

### Piped Input

Ghost accepts piped input, turning it into a data processing pipeline for your
neural interface. Stream files, command output, or net traffic directly to your
 AI companion.

```bash
# Jack in with inline code
echo "func decrypt(key []byte) error" | ghost "analyze this encryption routine"

# Scan system logs for intrusions
cat /var/log/auth.log | ghost "detect any unauthorized access attempts"

# Pull intel from the net
curl -s https://api.github.com/users/torvalds | ghost "profile this netrunner"
```

**Note:** Input stream limited to 10 MB. For larger data dumps, filter with `head`
 or `tail` before jacking in.

### Image Analysis

Ghost can analyze images using vision capable models, turning visual data into
 actionable intelligence:

```bash
# Analyze a single image
ghost --image "screenshot.png" "What security vulnerabilities do you see?"

# Analyze multiple images
ghost --image "diagram1.jpg" --image "diagram2.png" "Compare these network architectures"

# Combine with piped input
cat network-config.txt | ghost --image "topology.png" "Analyze this network setup"
```

**Note:** Image analysis requires a vision-capable model (default: `qwen2.5vl:7b`)
 and the images must be accessible from your local filesystem.

### Health Check

Run diagnostics to verify Ghost is properly configured and connected:

```bash
ghost health
```

The health command performs comprehensive system checks:

- **System Configuration**: Displays config file status, host, chat model,
 vision model, and system prompts
- **Neural Link Status**: Verifies Ollama API connectivity and version
- **Model Validation**: Confirms both chat and vision models are loaded and available

The command reports the status of both chat and vision models, Ollama API
connectivity, and displays the current configuration. If issues are detected,
it provides actionable error messages.

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
- `--model`: LLM to use for basic chat (default: `llama3.1:8b`)
- `--system`: System prompt override for basic chat model (optional)
- `--vision-model`: LLM to use for analyzing images (default: `qwen2.5vl:7b`)
- `--vision-system`: System prompt override for vision model (optional)
- `--vision-prompt`: Prompt for image analysis (default: "Analyze the attached
 image(s)")
- `--image`: Path to an image file (can be used multiple times)

### Configuration File

Create a config file at `~/.config/ghost/config.toml`:

```toml
host = "http://localhost:11434"
model = "llama3.1:8b"
system = "You are Ghost, a cyberpunk inspired terminal based assistant."

[vision]
model = "qwen2.5vl:7b"
system_prompt = ""
prompt = "Analyze the attached image(s)"
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
