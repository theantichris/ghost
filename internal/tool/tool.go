package tool

import (
	"context"
	"encoding/json"

	"github.com/theantichris/ghost/internal/llm"
)

// Tool is the interface that all the tools the LLM uses must implements.
type Tool interface {
	Definition() llm.Tool
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}
