package llm

import "errors"

var (
	// ErrValidation indicates input validation failure for LLM requests.
	ErrValidation = errors.New("failed to validate input")

	// ErrRequest indicates HTTP request creation or marshaling failure.
	ErrRequest = errors.New("failed to create request")

	// ErrResponse indicates HTTP response processing or parsing failure.
	ErrResponse = errors.New("failed to process response")
)
