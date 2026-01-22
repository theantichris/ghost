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
[![Go ReportCard](https://goreportcard.com/badge/theantichris/ghost)](https://goreportcard.com/report/theantichris/ghost)
![license](https://img.shields.io/badge/license-MIT-informational?style=flat)

> *Your personal AI companion in the terminal. Always on. Always local. No corp surveillance.*

Ghost is a command-line AI assistant powered by [Ollama](https://ollama.ai),
bringing the spirit of cyberpunk AI companions from *Shadowrun*, *Cyberpunk 2077*,
and *The Matrix* into your daily workflow.

## Jack In

**Prerequisites:**

- [Ollama](https://ollama.ai) installed and running
- At least one model pulled (e.g., `ollama pull llama3`)

**Install:**

```bash
go install github.com/theantichris/ghost@latest
```

Or grab prebuilt binaries from the [releases page](https://github.com/theantichris/ghost/releases).

**Run your first query:**

```bash
ghost "Explain recursion in simple terms"
```

## Core Capabilities

- **Intelligence on demand:** Ask questions, get explanations, analyze data
- **Interactive chat:** Conversation mode with Vim-style keybindings
- **Data stream analysis:** Pipe logs, files, or any text directly into Ghost
- **Visual recon:** Feed images to vision models for analysis and description
- **Format flexibility:** Output as plain text, JSON, or styled Markdown
- **Web search:** Real-time web searches via Tavily API when the model needs
 current information

## Usage Examples

```bash
# Query the net for intel
ghost "how do I crack open a encrypted data stream?"

# Scan your logs for anomalies
cat system.log | ghost "what's lurking in here?"

# Extract structured data
ghost "give me a list of common netrunner tools" -f json | jq .

# Generate formatted dossiers
ghost "write a guide to bypassing corp firewalls" -f markdown > intel.md

# Visual recon (requires vision model)
ghost "analyze this security feed" -i camera-feed.png

# Compare surveillance data
ghost "what changed in the facility?" -i before-raid.png -i after-raid.png

# Real-time intel (requires Tavily API key)
ghost "what are the latest vulnerabilities disclosed this week?"
```

## Interactive Chat

Launch a persistent conversation session with Ghost:

```bash
ghost chat
ghost chat --model llama3
```

**Vim-style keybindings:**

| Key      | Action                           |
|----------|----------------------------------|
| `i`      | Enter insert mode (start typing) |
| `Esc`    | Return to normal mode            |
| `Enter`  | Send message (in insert mode)    |
| `j`      | Scroll down one line             |
| `k`      | Scroll up one line               |
| `Ctrl+d` | Scroll down half page            |
| `Ctrl+u` | Scroll up half page              |
| `gg`     | Go to top                        |
| `G`      | Go to bottom                     |
| `:q`     | Disconnect from Ghost            |

## System Configuration

Configure Ghost via command-line flags, environment variables, or config file.

### Command Flags

- `-m, --model`: Model to use (e.g., `llama3`)
- `-V, --vision-model`: Vision model for images (defaults to main model)
- `-i, --image`: Image file path (can be used multiple times)
- `-f, --format`: Output format: `text`, `json`, or `markdown`
- `-u, --url`: Ollama API URL (default: `http://localhost:11434/api`)
- `-c, --config`: Config file path (default: `~/.config/ghost/config.toml`)

### Environment Variables

```bash
export GHOST_MODEL=llama3
export GHOST_VISION_MODEL=llama3.2-vision
export GHOST_URL=http://localhost:11434/api
export GHOST_SEARCH_API_KEY=tvly-xxxxx   # Tavily API key for web search
```

### Config File

Create `~/.config/ghost/config.toml`:

```toml
model = "llama3"
url = "http://localhost:11434/api"

[vision]
model = "llama3.2-vision"

[search]
api-key = "tvly-xxxxx"  # Get your key at tavily.com
max-results = 5         # Number of search results (default: 5)
```

## License

MIT
