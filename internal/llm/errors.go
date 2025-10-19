package llm

import (
	"errors"

	"github.com/theantichris/ghost/internal/exitcode"
)

var (
	// ErrNoHostURL indicates the Ollama host URL was not provided or is empty.
	ErrNoHostURL = exitcode.New(errors.New("no base URL provided"), exitcode.ExConfig)

	// ErrNoDefaultModel indicates the default model name was not provided or is empty.
	ErrNoDefaultModel = exitcode.New(errors.New("no default model provided"), exitcode.ExConfig)

	// ErrOllama indicates the Ollama API failed to return a valid response.
	ErrOllama = exitcode.New(errors.New("failed to get API response"), exitcode.ExUnavailable)
)
