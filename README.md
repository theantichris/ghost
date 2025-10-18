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

Ghost is a local, AI assistant CLI tool built in Go and powered by Ollama.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_,
_Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always-on AI
companion into a terminal-first experience.

## Prerequisites

- [Ollama](https://ollama.ai) installed and running
- Ollama model `dolphin-mixtral:8x7b` pulled and available

## Installation

### Via Go Install

```bash
go install github.com/theantichris/ghost@latest
```

### Pre-built Binaries

Download the latest release for your platform from the
[releases page](https://github.com/theantichris/ghost/releases).

## Usage

Invoke Ghost with a prompt to get instant AI assistance in your terminal:

```bash
ghost "your prompt here"
```

### Examples

```bash
# Get intel on tech
ghost "Explain how neural interfaces work"

# Quick data lookup
ghost "What's the difference between a netrunner and a decker?"

# Code assistance for your next run
ghost "Write a Go function to encrypt data with AES-256"
```

### Help

```bash
ghost --help
```

## Configuration

Ghost currently connects to Ollama at `http://100.92.199.66:11434` and uses the
`dolphin-mixtral:8x7b` model. Configuration options are planned for future
releases.

## License

MIT
