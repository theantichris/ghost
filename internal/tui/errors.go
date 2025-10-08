package tui

import "errors"

var (
	// ErrLLMClientInit indicates the LLM client was not properly initialized before use.
	ErrLLMClientInit = errors.New("LLM client not initialized")

	// ErrLLMRequest indicates an LLM request processing failure in the TUI context.
	ErrLLMRequest = errors.New("failed to process LLM request")
)
