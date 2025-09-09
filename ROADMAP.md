# Implementation Roadmap

## Phase 1: Foundation (Current)

- [ ] Basic CLI chat with streaming
- [ ] TUI scaffolding (Bubble Tea): single-window chat view with streaming render, input line, scrollback, basic keymap

## Phase 2: Core Tools

- [ ] Card system for prompts
- [ ] Tool framework architecture
- [ ] Card runner
- [ ] Web search capability
- [ ] Personal assistant; calendar, email, task, reminders
- [ ] Image generation (txt2img, simple img2img) via local engine
- [ ] TUI: command palette & status bar (e.g., / commands, tool-call progress, error toasts)
  - e.g., SDXL/FLUX running locally (InvokeAI/ComfyUI API or direct lib)
  - Minimal params: prompt, negative, steps, guidance, seed, size
  - Save outputs to a local gallery + metadata (prompt, seed)

## Phase 3: Memory & Knowledge

- [ ] Persistent chat history
- [ ] RAG implementation with vector database
- [ ] Document ingestion pipeline
- [ ] Basic conversation analysis
- [ ] Tool usage pattern learning
- [ ] Image asset indexing (store thumbnails + prompt metadata; tag for recall)
- [ ] TUI: conversation switcher & memory inspector (pane/tab to view sessions, search history, peek RAG hits)

## Phase 4: Advanced AI Features

- [ ] Personality evolution system
- [ ] Self-awareness mechanisms
- [ ] Knowledge forgetting system
- [ ] Advanced image editing (in painting/out painting, upscaling)
- [ ] ControlNet/conditioning (pose, depth, edges)
- [ ] Prompt auto-crafting from memory (style/subject preferences)
- [ ] TUI: persona manager & settings (select/preview Cards, per-assistant indicators)

## Phase 5: Future Expansion

- [ ] File system operations (with safety controls)
- [ ] Local code execution (sandboxed)
- [ ] Document processing and modification
- [ ] Secure file management
- [ ] Cross-device accessibility
- [ ] Enhanced user experience
- [ ] Personalization (LoRA/DreamBooth on your own photos/styles)
- [ ] Batch/pipeline workflows (ComfyUI graphs, queued jobs, scheduler)
- [ ] TUI: theming & layout presets (Lip Gloss styling, detachable panes, export logs)
