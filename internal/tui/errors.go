package tui

import "errors"

var (
	ErrLLMClientInit = errors.New("LLM client not initialized")
	ErrLLMRequest    = errors.New("failed to process LLM request")
)
