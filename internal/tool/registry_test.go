package tool

import (
	"context"
	"encoding/json"
	"errors"
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

func TestDefinitions(t *testing.T) {
	registry := NewRegistry()

	tool := mockTool{
		name:   "mock tool",
		result: "mock result",
	}

	registry.Tools[tool.Definition().Function.Name] = tool

	got := registry.Definitions()

	if len(got) == 0 {
		t.Fatal("Definitions() count = 0, want 1")
	}

	if got[0].Function.Name != tool.Definition().Function.Name {
		t.Fatalf("Definitions() name = %s, want %s", got[0].Function.Name, tool.Definition().Function.Name)
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name       string
		tool       mockTool
		calledTool string
		wantErr    bool
		err        error
	}{
		{
			name:       "executes tool",
			tool:       mockTool{name: "mock tool", result: "mock result"},
			calledTool: "mock tool",
		},
		{
			name:       "returns error for unknown tool",
			tool:       mockTool{name: "mock tool", result: "mock result"},
			calledTool: "unknown tool",
			wantErr:    true,
			err:        ErrToolNotRegistered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			registry.Tools[tt.tool.Definition().Function.Name] = tt.tool

			got, err := registry.Execute(context.Background(), tt.calledTool, json.RawMessage{})

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("Execute() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Execute() err = %v, want nil", err)
			}

			if got != tt.tool.result {
				t.Errorf("Execute() got = %s, want %s", got, tt.tool.result)
			}
		})
	}
}
