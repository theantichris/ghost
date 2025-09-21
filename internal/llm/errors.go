package llm

import "errors"

var (
	ErrURLEmpty          = errors.New("baseURL cannot be empty")
	ErrModelEmpty        = errors.New("defaultModel cannot be empty")
	ErrChatHistoryEmpty  = errors.New("chat history cannot be empty")
	ErrMarshalRequest    = errors.New("failed to marshal request body")
	ErrCreateRequest     = errors.New("failed to create HTTP request")
	ErrClientResponse    = errors.New("failed to get response from Ollama API")
	ErrNon2xxResponse    = errors.New("non-2xx response")
	ErrResponseBody      = errors.New("failed to read response body")
	ErrUnmarshalResponse = errors.New("failed to unmarshal response body")
	ErrTimeout           = errors.New("request to LLM timed out")
)
