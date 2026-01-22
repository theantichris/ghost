package ui

import (
	"context"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
)

func newTestModel() ChatModel {
	logger := log.New(io.Discard)

	return NewChatModel(context.Background(), "http://localhost:11434/api", "test-model", "test system", logger)
}

func TestChatModel_ModeTransitions(t *testing.T) {
	tests := []struct {
		name      string
		startMode Mode
		msg       tea.Msg
		wantMode  Mode
	}{
		{
			name:      "i enters insert mode from normal",
			startMode: ModeNormal,
			msg:       tea.KeyPressMsg{Code: 'i', Text: "i"},
			wantMode:  ModeInsert,
		},
		{
			name:      "escape returns to normal from insert",
			startMode: ModeInsert,
			msg:       tea.KeyPressMsg{Code: tea.KeyEscape},
			wantMode:  ModeNormal,
		},
		{
			name:      "colon enters command mode from normal",
			startMode: ModeNormal,
			msg:       tea.KeyPressMsg{Code: ':', Text: ":"},
			wantMode:  ModeCommand,
		},
		{
			name:      "escape returns to normal from command",
			startMode: ModeCommand,
			msg:       tea.KeyPressMsg{Code: tea.KeyEscape},
			wantMode:  ModeNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = tt.startMode
			model.ready = true

			newModel, _ := model.Update(tt.msg)
			got := newModel.(ChatModel).mode

			if got != tt.wantMode {
				t.Errorf("mode = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

func TestChatModel_QuitCommand(t *testing.T) {
	model := newTestModel()
	model.mode = ModeCommand
	model.cmdBuffer = "q"
	model.ready = true

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestChatModel_InvalidCommand(t *testing.T) {
	model := newTestModel()
	model.mode = ModeCommand
	model.cmdBuffer = "invalid"
	model.ready = true

	newModel, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := newModel.(ChatModel)

	if got.mode != ModeNormal {
		t.Errorf("mode = %v, want ModeNormal", got.mode)
	}

	if got.cmdBuffer != "" {
		t.Errorf("cmdBuffer = %q, want empty", got.cmdBuffer)
	}
}

func TestChatModel_LLMResponseMsg(t *testing.T) {
	model := newTestModel()
	model.ready = true
	model.responseCh = make(chan tea.Msg)

	newModel, cmd := model.Update(LLMResponseMsg("hello "))

	got := newModel.(ChatModel)

	if got.history != "hello " {
		t.Errorf("history = %q, want %q", got.history, "hello ")
	}

	if got.currentResponse != "hello " {
		t.Errorf("currentResponse = %q, want %q", got.currentResponse, "hello ")
	}

	if cmd == nil {
		t.Error("expected listenForChunk command, got nil")
	}
}

func TestChatModel_LLMDoneMsg(t *testing.T) {
	model := newTestModel()
	model.ready = true
	model.currentResponse = "test response"
	model.history = "You: hi\n\nghost: test response"

	newModel, _ := model.Update(LLMDoneMsg{})

	got := newModel.(ChatModel)

	if got.currentResponse != "" {
		t.Errorf("currentResponse = %q, want empty", got.currentResponse)
	}

	if len(got.messages) != 2 {
		t.Errorf("messages length = %d, want 2", len(got.messages))
	}

	lastMsg := got.messages[len(got.messages)-1]
	if lastMsg.Role != llm.RoleAssistant {
		t.Errorf("last message role = %v, want RoleAssistant", lastMsg.Role)
	}

	if lastMsg.Content != "test response" {
		t.Errorf("last message content = %q, want %q", lastMsg.Content, "test response")
	}
}

func TestChatModel_LLMErrorMsg(t *testing.T) {
	model := newTestModel()
	model.ready = true

	testErr := io.EOF
	newModel, _ := model.Update(LLMErrorMsg{Err: testErr})

	got := newModel.(ChatModel)

	if got.history == "" {
		t.Error("expected error to be added to history")
	}
}

func TestChatModel_GGStateMachine(t *testing.T) {
	tests := []struct {
		name          string
		keys          []string
		wantAwaitingG bool
	}{
		{
			name:          "single g sets awaitingG",
			keys:          []string{"g"},
			wantAwaitingG: true,
		},
		{
			name:          "gg resets awaitingG",
			keys:          []string{"g", "g"},
			wantAwaitingG: false,
		},
		{
			name:          "g then other key resets awaitingG",
			keys:          []string{"g", "j"},
			wantAwaitingG: false,
		},
		{
			name:          "g then i resets awaitingG",
			keys:          []string{"g", "i"},
			wantAwaitingG: false,
		},
		{
			name:          "G resets awaitingG",
			keys:          []string{"g", "G"},
			wantAwaitingG: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = ModeNormal
			model.ready = true

			var result tea.Model = model
			for _, key := range tt.keys {
				result, _ = result.Update(tea.KeyPressMsg{Text: key})
			}

			got := result.(ChatModel).awaitingG
			if got != tt.wantAwaitingG {
				t.Errorf("awaitingG = %v, want %v", got, tt.wantAwaitingG)
			}
		})
	}
}

func TestChatModel_VimKeybindings(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "j scrolls down", key: "j"},
		{name: "k scrolls up", key: "k"},
		{name: "ctrl+d half page down", key: "ctrl+d"},
		{name: "ctrl+u half page up", key: "ctrl+u"},
		{name: "G goes to bottom", key: "G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = ModeNormal
			model.ready = true

			// Verify key is handled without error and stays in normal mode
			newModel, _ := model.Update(tea.KeyPressMsg{Text: tt.key})
			got := newModel.(ChatModel)

			if got.mode != ModeNormal {
				t.Errorf("mode = %v, want ModeNormal", got.mode)
			}
		})
	}
}
