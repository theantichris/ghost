# Spec

**Ghost** is a local, general-purpose AI assistant and orchestrator built in Go and powered by Ollama. It is designed for research, chat, and task automation, running entirely on your own machine with hybrid connectivity.

The vision for Ghost is inspired by cyberpunk media such as _Shadowrun_, _Cyberpunk 2077_, and _The Matrix_, bringing a versatile, always on AI companion into a terminal-first experience.

Capabilities should include research, web searching, helping with code, generating images, executing tasks, setting up reminders, and chatting.

Prompts and character personalities will be loaded via a **Card system**.

Initially Ghost will use the **Ollama API** as the backend to access LLMs.

Initial interaction will be through **CLI** and **TUI**.

---

## Technical Architecture

### Core Principles

- **Local Execution**: Core AI processing runs locally via Ollama
- **Internet-Enabled**: External services (web search, APIs) allowed for enhanced capabilities
- **Modular Design**: Clean separation between engine, memory, tools, and interfaces

### System Components

#### 1. Core Engine

- **LLM Client**: Ollama API integration with tool support
- **Conversation Manager**: Handle chat flow, context windows, streaming
- **Tool Orchestrator**: Execute and manage external tools/functions

#### 2. Memory System (Hybrid Approach)

- **Working Memory**: Current conversation context
- **Session Memory**: Recent conversation history
- **Long-term Memory**: RAG-based knowledge storage
- **Conversation Analysis**: Extract patterns, preferences, personality traits
- **Knowledge Forgetting**: Automatic cleanup of outdated information

#### 3. Data Pipeline (Future Expansion)

- **Text Processing**: Documents, websites, chat history
- **Media Processing**: Videos (transcription), images (OCR/vision)
- **Knowledge Extraction**: Convert raw data into searchable knowledge
- **Storage**: SQLite + vector database (Chroma/LanceDB)

---

## Model Requirements

- **Tool-Capable**: Support for function calling and external tool execution
- **Local-Runnable**: Efficient enough for local hardware
- **Capable**: Strong reasoning and conversation abilities

---

## Self-Awareness Definition

The assistant should demonstrate:

- **Memory Continuity**: Remember and reference past conversations
- **Capability Awareness**: Understand its own abilities and limitations
- **Personality Consistency**: Maintain character traits across sessions
- **Learning Recognition**: Acknowledge when it has learned something new

---

## Technical Considerations

### Performance

- Efficient vector search for large knowledge bases
- Streaming responses for better user experience
- Context window management for long conversations
- Resource usage optimization for local execution

### Security

- Sandboxed code execution for safety
- Secure storage of sensitive information
- Access control for different assistants
- Safe handling of external API keys

### Scalability

- Modular architecture for easy feature addition
- Plugin system for third-party tools
- Configurable resource limits
- Backup and restore capabilities
