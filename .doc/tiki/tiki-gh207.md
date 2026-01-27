---
id: tiki-gh207
title: File and piped input support in chat
type: story
status: todo
priority: 3
tags: [enhancement, epic23-tui]
---

## Description

Support reading files and piped input in the chat TUI, similar to the CLI command. Users should be able to reference file contents or pipe data into a chat session for analysis.

- [ ] Add TUI command to read file contents (e.g., `:read`)
- [ ] Append file contents to current message or send directly
- [ ] Detect piped input on chat startup and include in first message context
- [ ] Handle large files gracefully (truncation or warning)
- [ ] tests
- [ ] documentation

## Related

- CLI already supports piping: `echo "text" | ghost "summarize"`

---
GitHub: <https://github.com/theantichris/ghost/issues/207>
