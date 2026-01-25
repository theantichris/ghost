package tool

import (
	"context"
	"encoding/json"

	"github.com/theantichris/ghost/v3/internal/llm"
)

// MockTool is a test double for the Tool interface.
type MockTool struct {
	Name   string
	Result string
	Err    error
}

func (t MockTool) Definition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        t.Name,
			Description: "mock tool",
			Parameters:  llm.ToolParameters{Type: "object"},
		},
	}
}

func (t MockTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	return t.Result, t.Err
}
