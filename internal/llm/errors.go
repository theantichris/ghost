package llm

import (
	"errors"

	"github.com/theantichris/ghost/internal/exitcode"
)

var (
	// ErrNoHostURL indicates the Ollama host URL was not provided or is empty.
	ErrNoHostURL = exitcode.New(errors.New("no base URL provided"), exitcode.ExConfig)

	// ErrOllama indicates the Ollama API failed to return a valid response.
	ErrOllama = exitcode.New(errors.New("failed to get API response"), exitcode.ExUnavailable)

	// ErrModelNotFound indicates the requested model is not available.
	ErrModelNotFound = exitcode.New(errors.New("model not found"), exitcode.ExUnavailable)
)
