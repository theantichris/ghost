package ui

import (
	"context"
	"errors"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/tool"
)

func newTestCLIModel(t *testing.T) CLIModel {
	t.Helper()

	logger := log.New(io.Discard)
	registry := tool.NewRegistry("", 0, logger)

	config := ModelConfig{
		Context:  context.Background(),
		URL:      "http://localhost/11434/api",
		ChatLLM:  "test-model",
		Prompts:  agent.Prompt{System: "test system prompt"},
		Registry: registry,
		Logger:   logger,
	}

	model, err := NewCLIModel(config, "test prompt")
	if err != nil {
		t.Fatal(err)
	}

	return model
}

func TestCLIModel_Update(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		wantContent string
		wantDone    bool
		wantErr     bool
		wantCmd     bool
		wantQuit    bool
	}{
		{
			name:        "StreamChunkMsg accumulates content and returns listen command",
			msg:         StreamChunkMsg("hello "),
			wantContent: "hello ",
			wantDone:    false,
			wantCmd:     true,
			wantQuit:    false,
		},
		{
			name:        "LLMDoneMsg sets done and returns quit",
			msg:         LLMDoneMsg{},
			wantContent: "",
			wantDone:    true,
			wantCmd:     true,
			wantQuit:    true,
		},
		{
			name:        "StreamErrorMsg sets error and done, returns quit",
			msg:         StreamErrorMsg{Err: errors.New("test error")},
			wantContent: "",
			wantDone:    true,
			wantErr:     true,
			wantCmd:     true,
			wantQuit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestCLIModel(t)

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(CLIModel)

			if got.content != tt.wantContent {
				t.Errorf("content = %q, want %q", got.content, tt.wantContent)
			}

			if got.done != tt.wantDone {
				t.Errorf("done = %v, want %v", got.done, tt.wantDone)
			}

			if tt.wantErr && got.Err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && got.Err != nil {
				t.Errorf("expected no error, got %v", got.Err)
			}

			if tt.wantCmd && cmd == nil {
				t.Error("expected command, got nil")
			}

			if !tt.wantCmd && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}
		})
	}
}

func TestCLIModel_ChunkAccumulation(t *testing.T) {
	model := newTestCLIModel(t)

	chunks := []string{"Hello", " ", "world", "!"}
	for _, chunk := range chunks {
		newModel, _ := model.Update(StreamChunkMsg(chunk))
		model = newModel.(CLIModel)
	}

	want := "Hello world!"
	if model.content != want {
		t.Errorf("accumulated content = %q, want %q", model.content, want)
	}
}
