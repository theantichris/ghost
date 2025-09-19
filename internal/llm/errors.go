package llm

import "errors"

var ErrURLEmpty = errors.New("baseURL cannot be empty")
var ErrModelEmpty = errors.New("defaultModel cannot be empty")
var ErrMessageEmpty = errors.New("message cannot be empty")
var ErrMarshalRequest = errors.New("failed to marshal request body")
var ErrCreateRequest = errors.New("failed to create HTTP request")
var ErrClientResponse = errors.New("failed to get response from Ollama API")
var ErrNon2xxResponse = errors.New("non-2xx response")
var ErrResponseBody = errors.New("failed to read response body")
var ErrUnmarshalResponse = errors.New("failed to unmarshal response body")
