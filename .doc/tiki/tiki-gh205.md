---
id: tiki-gh205
title: Refactor root command to use channel-based streaming pattern
type: story
status: todo
priority: 3
tags: [enhancement, epic23-tui]
---

## Description

The root command currently uses `Program.Send()` from a goroutine to stream LLM responses. The new chat command implements a more idiomatic Bubbletea pattern using channels and recursive commands.

## Current (root command)

- Goroutine calls `StreamChat` with callback
- Callback uses `program.Send()` to push chunks
- Requires passing program reference

## Proposed (chat command pattern)

- Goroutine sends chunks to a channel
- `listenForChunk` command waits for one chunk, returns it as a message
- `Update()` handles message and returns `listenForChunk` again
- Channel close signals completion

## Benefits

- More idiomatic TEA pattern (pure message passing)
- No shared program reference needed
- Easier to test
- Consistent patterns across codebase

## Related

- Implemented in chat command: #180

---
GitHub: <https://github.com/theantichris/ghost/issues/205>
