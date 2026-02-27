package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestListenForChunk(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(ch chan tea.Msg)
		wantType string
		wantMsg  tea.Msg
	}{
		{
			name: "returns message from channel",
			setup: func(ch chan tea.Msg) {
				ch <- StreamChunkMsg("hello")
			},
			wantType: "StreamChunkMsg",
			wantMsg:  StreamChunkMsg("hello"),
		},
		{
			name: "returns LLMDoneMsg when channel closes",
			setup: func(ch chan tea.Msg) {
				close(ch)
			},
			wantType: "LLMDoneMsg",
			wantMsg:  LLMDoneMsg{},
		},
		{
			name: "returns error message from channel",
			setup: func(ch chan tea.Msg) {
				ch <- StreamErrorMsg{Err: nil}
			},
			wantType: "StreamErrorMsg",
			wantMsg:  StreamErrorMsg{Err: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan tea.Msg, 1)
			tt.setup(ch)

			cmd := listenForChunk(ch)
			if cmd == nil {
				t.Fatal("listenForChunk() returned nil command")
			}

			got := cmd()

			switch want := tt.wantMsg.(type) {
			case StreamChunkMsg:
				gotChunk, ok := got.(StreamChunkMsg)
				if !ok {
					t.Fatalf("got %T, want StreamChunkMsg", got)
				}
				if gotChunk != want {
					t.Errorf("got %q, want %q", gotChunk, want)
				}

			case LLMDoneMsg:
				if _, ok := got.(LLMDoneMsg); !ok {
					t.Fatalf("got %T, want LLMDoneMsg", got)
				}

			case StreamErrorMsg:
				if _, ok := got.(StreamErrorMsg); !ok {
					t.Fatalf("got %T, want StreamErrorMsg", got)
				}
			}
		})
	}
}
