package tool

import (
	"context"
	"encoding/json"

	"github.com/theantichris/ghost/v3/internal/llm"
)

// Tool is the interface that all the tools the LLM uses must implement.
type Tool interface {
	Definition() llm.Tool
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}
