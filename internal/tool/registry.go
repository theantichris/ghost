package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
)

var ErrToolNotRegistered = errors.New("tool not registered")

// Registry holds all available tools, provides their definitions to send to chat
// requests, and dispatches execution to the right tool by name.
type Registry struct {
	Tools map[string]Tool
}

// NewRegistry creates a new Registry and initializes the tool map.
func NewRegistry(tavilyAPIKey string, maxResults int, logger *log.Logger) Registry {
	registry := Registry{
		Tools: map[string]Tool{},
	}

	if tavilyAPIKey != "" {
		if maxResults == 0 {
			maxResults = 5
		}

		registry.Register(NewSearch(tavilyAPIKey, maxResults))
		logger.Debug("tool registered", "name", "web_search")
	}

	return registry
}

// Register adds a tool to the registry using its definition's name as the key.
func (registry *Registry) Register(tool Tool) {
	registry.Tools[tool.Definition().Function.Name] = tool
}

// Definitions returns a slice of all tool definitions.
func (registry *Registry) Definitions() []llm.Tool {
	var definitions []llm.Tool

	for _, tool := range registry.Tools {
		definitions = append(definitions, tool.Definition())
	}

	return definitions
}

// Execute looks up the tool by name, calls its Execute() function, and returns
// the result.
// Returns an error if the tool isn't found.
func (registry *Registry) Execute(ctx context.Context, name string, args json.RawMessage) (string, error) {
	tool, ok := registry.Tools[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrToolNotRegistered, name)
	}

	result, err := tool.Execute(ctx, args)

	return result, err
}
