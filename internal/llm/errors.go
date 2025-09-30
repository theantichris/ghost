package llm

import "errors"

var (
	ErrValidation = errors.New("failed to validate input")
	ErrRequest    = errors.New("failed to create request")
	ErrResponse   = errors.New("failed to process response")
)
