package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

type mockTool struct {
	name   string
	result string
	err    error
}

func (tool mockTool) Definition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        tool.name,
			Description: "mock tool",
			Parameters:  llm.ToolParameters{Type: "object"},
		},
	}
}

func (tool mockTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	return tool.result, tool.err
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()

	tool := mockTool{
		name:   "mock tool",
		result: "mock result",
	}

	registry.Register(tool)

	_, ok := registry.Tools["mock tool"]
	if !ok {
		t.Errorf("Register() failed to register mock tool")
	}
}
