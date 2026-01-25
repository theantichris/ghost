package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	registry := NewRegistry()

	tool := MockTool{
		Name:   "mock tool",
		Result: "mock result",
	}

	registry.Register(tool)

	_, ok := registry.Tools["mock tool"]
	if !ok {
		t.Errorf("Register() failed to register mock tool")
	}
}

func TestDefinitions(t *testing.T) {
	registry := NewRegistry()

	tool := MockTool{
		Name:   "mock tool",
		Result: "mock result",
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
		tool       MockTool
		calledTool string
		wantErr    bool
		err        error
	}{
		{
			name:       "executes tool",
			tool:       MockTool{Name: "mock tool", Result: "mock result"},
			calledTool: "mock tool",
		},
		{
			name:       "returns error for unknown tool",
			tool:       MockTool{Name: "mock tool", Result: "mock result"},
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

			if got != tt.tool.Result {
				t.Errorf("Execute() got = %s, want %s", got, tt.tool.Result)
			}
		})
	}
}
