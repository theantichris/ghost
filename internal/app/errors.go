package app

import "errors"

var ErrLLMClientNil = errors.New("llmClient cannot be nil")
var ErrChatFailed = errors.New("chat failed")
