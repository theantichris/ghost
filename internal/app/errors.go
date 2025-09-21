package app

import "errors"

var (
	ErrLLMClientNil   = errors.New("llmClient cannot be nil")
	ErrChatFailed     = errors.New("chat failed")
	ErrReadingInput   = errors.New("error reading input")
	ErrUserInputEmpty = errors.New("user input empty")
)
