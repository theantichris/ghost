package tool

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theantichris/ghost/internal/llm"
)

var (
	ErrNoAPIKey     = errors.New("tavily API key not found")
	ErrSearchFailed = errors.New("matrix search failed")
)

// Search holds the web search configuration.
type Search struct {
	APIKey     string
	MaxResults int
}

// NewSearch creates and returns a new search config.
func NewSearch(apiKey string, maxResults int) Search {
	search := Search{
		APIKey:     apiKey,
		MaxResults: maxResults,
	}

	return search
}

// Definition returns the tool schema.
func (search Search) Definition() llm.Tool {
	parameters := llm.ToolParameters{
		Type:     "object",
		Required: []string{"query"},
		Properties: map[string]llm.ToolProperty{
			"query": {
				Type:        "string",
				Description: "the search query",
			},
		},
	}

	searchTool := llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "web_search",
			Description: "search the web for current information, news, and real time data",
			Parameters:  parameters,
		},
	}

	return searchTool
}

func (search Search) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	// Parse arguments into a struct with Query string
	// Call Tavily API
	// Format results as a string and return

	return "", nil
}
