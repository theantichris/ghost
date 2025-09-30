package llm

import "errors"

var (
	ErrURLEmpty          = errors.New("failed to get api url")
	ErrModelEmpty        = errors.New("failed to get model name")
	ErrChatHistoryEmpty  = errors.New("failed to get chat history")
	ErrMarshalRequest    = errors.New("failed to marshal request body")
	ErrCreateRequest     = errors.New("failed to create HTTP request")
	ErrClientResponse    = errors.New("failed to get response from Ollama API")
	ErrNon2xxResponse    = errors.New("non-2xx response")
	ErrResponseBody      = errors.New("failed to read response body")
	ErrUnmarshalResponse = errors.New("failed to unmarshal response body")
	ErrTimeout           = errors.New("request to LLM timed out")
)
