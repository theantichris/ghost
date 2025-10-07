# TUI Implementation Progress

This document tracks the implementation of Issue #9: TUI scaffolding.

## Completed Steps

### Step 1: Add BubbleTea Dependency

- ✅ Added `github.com/charmbracelet/bubbletea` to project dependencies

### Step 2: Define TUI Model Structure

- ✅ Created `Model` struct with all necessary fields:
  - Dependencies: `logger`, `llmClient`
  - Chat data: `chatHistory`, `messages`
  - UI state: `input`, `width`, `height`
  - Streaming state: `streaming`, `currentMsg`
  - Exit state: `exiting`, `err`

### Step 3: Handle Terminal Window Messages (TDD)

- ✅ Test: Window size message updates model dimensions
- ✅ Implementation: `Update()` handles `tea.WindowSizeMsg`

### Step 4: Handle Regular Key Press (TDD)

- ✅ Test: Typing appends characters to input
- ✅ Implementation: Handle `tea.KeyRunes` message

### Step 5: Handle Backspace Key (TDD)

- ✅ Test: Backspace removes last character from input
- ✅ Implementation: Handle `tea.KeyBackspace` with safety check

### Step 6: Handle Ctrl+D to Exit (TDD)

- ✅ Test: Ctrl+D sets exiting flag and returns quit command
- ✅ Implementation: Handle `tea.KeyCtrlD` returning `tea.Quit`

### Step 7: Handle Ctrl+C to Exit (TDD)

- ✅ Test: Ctrl+C sets exiting flag and returns quit command
- ✅ Implementation: Handle both `tea.KeyCtrlD` and `tea.KeyCtrlC`
- ✅ Refactored to nested switch for cleaner key handling

### Step 8: Handle Enter Key to Clear Input (TDD)

- ✅ Test: Enter key clears input field
- ✅ Implementation: Handle `tea.KeyEnter` with empty input check

### Step 9: Add Message to Chat History on Enter (TDD)

- ✅ Test: Enter adds user message to chat history
- ✅ Implementation: Append `ChatMessage` with `UserRole`

### Step 10: Implement Basic View Rendering (TDD)

- ✅ Test: View contains input field
- ✅ Implementation: `View()` returns input string

### Step 11: Render Separator Line (TDD)

- ✅ Test: View contains horizontal separator
- ✅ Implementation: Add separator using box-drawing character

### Step 12: Render Chat Messages (TDD)

- ✅ Test: View contains chat messages from messages array
- ✅ Implementation: Join messages and display above separator

### Step 13: Handle /bye and /exit Commands (TDD)

- ✅ Test: /bye command sets exiting and replaces with "Goodbye!"
- ✅ Implementation: Check for exit commands and transform message

### Step 14: Create Model Constructor (TDD)

- ✅ Test: NewModel initializes with dependencies
- ✅ Implementation: Constructor with logger parameter

### Step 15: Initialize with System Prompt (TDD)

- ✅ Test: NewModel initializes chat history with system prompt
- ✅ Implementation: Add system prompt to initial chat history

### Step 16: Add Greeting Prompt to Constructor (TDD)

- ✅ Test: NewModel adds "Greet the user" to chat history
- ✅ Implementation: Initialize with system prompt + greeting instruction

### Step 17: Trigger LLM Request on Init (TDD)

- ✅ Test: Init returns command when chat history exists
- ✅ Implementation: Return `sendChatRequest` method as command

### Step 18: Define Custom Message Types

- ✅ Created `streamChunkMsg` - carries token content
- ✅ Created `streamCompleteMsg` - signals streaming complete with accumulated content
- ✅ Created `streamErrorMsg` - carries error information

### Step 19: Handle Stream Chunk Messages (TDD)

- ✅ Test: Chunk messages append to currentMsg and set streaming flag
- ✅ Implementation: Handle `streamChunkMsg` in Update

### Step 20: Handle Stream Complete Messages (TDD)

- ✅ Test: Complete message adds to messages and chatHistory, resets state
- ✅ Implementation: Handle `streamCompleteMsg` in Update

### Step 21: Handle Stream Error Messages (TDD)

- ✅ Test: Error message stops streaming and stores error
- ✅ Implementation: Handle `streamErrorMsg` in Update

### Step 22: Implement sendChatRequest Method (TDD)

- ✅ Test: Calls LLM client with chat history
- ✅ Implementation: Call llmClient.Chat with context and callback
- ✅ Fixed: Changed `llmClient` from `*llm.LLMClient` to `llm.LLMClient`

### Step 23: Accumulate Tokens in sendChatRequest (TDD)

- ✅ Test: Tokens are accumulated and passed in complete message
- ✅ Implementation: Use `strings.Builder` to collect tokens
- ✅ Updated `streamCompleteMsg` to carry content

## Remaining Steps

### High Priority - Core Functionality

1. **Trigger LLM Request on User Message**
   - When user presses Enter (after adding to history), trigger LLM request
   - Return `sendChatRequest` command from Update on Enter key
   - Test that Enter with input triggers LLM call

2. **Handle Initial Greeting**
   - Make `Init()` actually call the LLM for the initial greeting
   - Ensure greeting response is displayed before user types
   - Test that model initializes and shows greeting

3. **Wire TUI into Chat Command**
   - Update `cmd/chat.go` to initialize Model with dependencies
   - Replace current chat loop with `tui.Run()`
   - Pass `logger`, `llmClient`, and `systemPrompt` to NewModel
   - Update `internal/tui/run.go` to accept and use Model

4. **Strip Think Blocks from Display**
   - Integrate `stdio.OutputWriter` think block filtering
   - Apply filtering when accumulating tokens in `sendChatRequest`
   - Test that `<think>...</think>` blocks are removed
   - Ensure filtered content is added to messages and chatHistory

### Medium Priority - UI Polish

1. **Display Streaming Indicator**
   - Show visual indicator when `streaming` is true
   - Could be a spinner, "..." animation, or status text
   - Update View to show indicator below input

2. **Display Errors in View**
   - Check if `model.err` is set in View
   - Display error message above input field
   - Style error messages for visibility

3. **Handle Scrolling**
   - Calculate visible lines based on `height`
   - Show most recent messages when chat exceeds screen height
   - Consider using a viewport component or manual slicing

4. **Improve Message Formatting**
   - Add visual distinction between user and assistant messages
   - Consider prefixes like "You: " and "Ghost: "
   - Handle word wrapping for long messages

### Low Priority - Enhancements

1. **Add True Live Streaming**
   - Current implementation accumulates all tokens before displaying
   - Enhance to show tokens as they arrive in real-time
   - Might require channel-based approach or tea.Batch

2. **Handle Context Cancellation**
   - Replace `context.Background()` with cancellable context
   - Allow Ctrl+C during streaming to cancel LLM request
   - Clean up resources on cancellation

3. **Add Loading State on Startup**
   - Show loading indicator while initial greeting generates
   - Disable input until first response complete

4. **Improve Test Coverage**
   - Add integration tests that test full flows
   - Test error scenarios more thoroughly
   - Add tests for View rendering edge cases

## Architecture Notes

### Current Design Decisions

- **Streaming**: Currently accumulates all tokens in `sendChatRequest`
  and returns complete message
  - Simpler implementation, works well for typical response times
  - Future: Could enhance with live token-by-token display

- **Message Types**: Custom message types for LLM streaming events
  - `streamChunkMsg` - individual tokens (currently not used)
  - `streamCompleteMsg` - complete response with accumulated content
  - `streamErrorMsg` - error handling

- **State Management**: Model holds all state, Update handles all events
  - Following standard BubbleTea pattern
  - Model is pass-by-value, must return updated model from Update

- **Interface Types**: `llmClient` is `llm.LLMClient` (interface), not pointer
  - Allows any implementation (real client or mock)
  - Follows Go best practices for interfaces

### Testing Strategy

- **TDD Approach**: Red-Green-Refactor cycle for all features
- **Table-Driven Tests**: Using `t.Run()` for test organization
- **Mocking**: Using `llm.MockLLMClient` for testing without API calls
- **Explicit Naming**: `actual` and `expected` prefixes in assertions

## References

- **Issue**: #9 - TUI scaffolding
- **Milestone**: basic TUI
- **Related Files**:
  - `internal/tui/model.go` - Model and Update logic
  - `internal/tui/model_test.go` - TDD tests
  - `internal/tui/run.go` - Entry point (needs updating)
  - `cmd/chat.go` - Command integration (needs updating)
  - `internal/llm/client.go` - LLM client interface
  - `internal/stdio/output.go` - Think block filtering

## Timeline

- **Started**: October 7, 2025
- **Current Phase**: Core TUI implementation complete, integration pending
- **Next Session**: Wire TUI into chat command and handle LLM requests
