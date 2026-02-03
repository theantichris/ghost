package ui

import (
	"context"
	"errors"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/tool"
)

func newTestStreamModel() StreamModel {
	logger := log.New(io.Discard)
	registry := tool.NewRegistry("", 0, logger)

	config := ModelConfig{
		Context:  context.Background(),
		URL:      "http://localhost/11434/api",
		Model:    "test-model",
		Registry: registry,
		Logger:   logger,
	}

	return NewStreamModel(config)
}

func TestStreamModel_Update(t *testing.T) {
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
			model := newTestStreamModel()

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(StreamModel)

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

func TestStreamModel_ChunkAccumulation(t *testing.T) {
	model := newTestStreamModel()

	chunks := []string{"Hello", " ", "world", "!"}
	for _, chunk := range chunks {
		newModel, _ := model.Update(StreamChunkMsg(chunk))
		model = newModel.(StreamModel)
	}

	want := "Hello world!"
	if model.content != want {
		t.Errorf("accumulated content = %q, want %q", model.content, want)
	}
}
