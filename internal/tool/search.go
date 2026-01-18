package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/theantichris/ghost/internal/llm"
)

const tavilyURL = "https://api.tavily.com/search"

var (
	ErrNoAPIKey     = errors.New("tavily API key not found")
	ErrSearchFailed = errors.New("matrix search failed")
	ErrParseArgs    = errors.New("failed to parse arguments")
)

type searchRequest struct {
	APIKey     string `json:"api_key"`
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type searchResults struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type searchResponse struct {
	Results []searchResults `json:"results"`
}

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

// Execute parses the arguments into the search query, makes a request to the
// Tavily API, then formats and returns the results as a string.
func (search Search) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var searchArgs struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal(args, &searchArgs); err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseArgs, err)
	}

	req := searchRequest{
		APIKey:     search.APIKey,
		Query:      searchArgs.Query,
		MaxResults: search.MaxResults,
	}

	var resp searchResponse

	err := requests.URL(tavilyURL).
		BodyJSON(&req).
		ToJSON(&resp).
		Fetch(ctx)

	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSearchFailed, err)
	}

	var sb strings.Builder
	for i, result := range resp.Results {
		fmt.Fprintf(&sb, "Result: %d: %s\n", i+1, result.Title)
		fmt.Fprintf(&sb, "URL: %s\n", result.URL)
		fmt.Fprintf(&sb, "%s\n\n", result.Content)
	}

	return sb.String(), nil
}
