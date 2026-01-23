---
id: tiki-gh208
title: Image analysis in chat
type: story
status: todo
priority: 3
tags: [enhancement]
---

## Description

Support image analysis in the chat TUI using vision models. Users should be able to share images for the LLM to analyze mid-conversation.

- [ ] Add TUI command to analyze image (e.g., `:image`)
- [ ] Use configured vision model for analysis
- [ ] Display image analysis results in conversation
- [ ] Support multiple images in sequence
- [ ] tests
- [ ] documentation

## Related

- CLI already supports `--image` flag for vision analysis
- Vision system uses separate model configured via `vision.model`

---
GitHub: <https://github.com/theantichris/ghost/issues/208>
