package ui

import (
	"context"
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
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
			msg:       tea.KeyPressMsg{Code: tea.KeyEscape, Text: "esc"},
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
			msg:       tea.KeyPressMsg{Code: tea.KeyEscape, Text: "esc"},
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

func TestChatModel_HandleInsertMode(t *testing.T) {
	tests := []struct {
		name                  string
		inputValue            string
		inputHistory          []string
		inputHistoryIndex     int
		msg                   tea.Msg
		wantMode              Mode
		wantInputValue        string
		wantMessageCount      int
		wantHistoryEmpty      bool
		wantInputHistory      []string
		wantInputHistoryIndex int
	}{
		{
			name:             "enter with empty input is ignored",
			inputValue:       "",
			msg:              tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"},
			wantMode:         ModeInsert,
			wantInputValue:   "",
			wantMessageCount: 1, // only system message
			wantHistoryEmpty: true,
		},
		{
			name:             "enter with whitespace input is ignored",
			inputValue:       "   ",
			msg:              tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"},
			wantMode:         ModeInsert,
			wantInputValue:   "   ",
			wantMessageCount: 1,
			wantHistoryEmpty: true,
		},
		{
			name:             "enter with content sends message",
			inputValue:       "hello ghost",
			msg:              tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"},
			wantMode:         ModeInsert,
			wantInputValue:   "",
			wantMessageCount: 2, // system + user
			wantHistoryEmpty: false,
		},
		{
			name:             "ctrl+j inserts newline",
			inputValue:       "line one",
			msg:              tea.KeyPressMsg{Code: 'j', Mod: tea.ModCtrl, Text: "ctrl+j"},
			wantMode:         ModeInsert,
			wantInputValue:   "line one\n",
			wantMessageCount: 1,
			wantHistoryEmpty: true,
		},
		{
			name:             "escape returns to normal mode",
			inputValue:       "some text",
			msg:              tea.KeyPressMsg{Code: tea.KeyEscape, Text: "esc"},
			wantMode:         ModeNormal,
			wantInputValue:   "some text",
			wantMessageCount: 1,
			wantHistoryEmpty: true,
		},
		{
			name:                  "up with empty history does nothing",
			inputValue:            "current text",
			inputHistory:          []string{},
			inputHistoryIndex:     0,
			msg:                   tea.KeyPressMsg{Code: tea.KeyUp},
			wantMode:              ModeInsert,
			wantInputValue:        "current text",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{},
			wantInputHistoryIndex: 0,
		},
		{
			name:                  "up navigates to previous history item",
			inputValue:            "",
			inputHistory:          []string{"first message", "second message"},
			inputHistoryIndex:     2, // at end of history
			msg:                   tea.KeyPressMsg{Code: tea.KeyUp},
			wantMode:              ModeInsert,
			wantInputValue:        "second message",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{"first message", "second message"},
			wantInputHistoryIndex: 1,
		},
		{
			name:                  "up at beginning of history stays at index 0",
			inputValue:            "first message",
			inputHistory:          []string{"first message", "second message"},
			inputHistoryIndex:     0, // already at beginning
			msg:                   tea.KeyPressMsg{Code: tea.KeyUp},
			wantMode:              ModeInsert,
			wantInputValue:        "first message",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{"first message", "second message"},
			wantInputHistoryIndex: 0,
		},
		{
			name:                  "down with empty history does nothing",
			inputValue:            "current text",
			inputHistory:          []string{},
			inputHistoryIndex:     0,
			msg:                   tea.KeyPressMsg{Code: tea.KeyDown},
			wantMode:              ModeInsert,
			wantInputValue:        "current text",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{},
			wantInputHistoryIndex: 0,
		},
		{
			name:                  "down navigates to next history item",
			inputValue:            "first message",
			inputHistory:          []string{"first message", "second message"},
			inputHistoryIndex:     0, // at first item
			msg:                   tea.KeyPressMsg{Code: tea.KeyDown},
			wantMode:              ModeInsert,
			wantInputValue:        "second message",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{"first message", "second message"},
			wantInputHistoryIndex: 1,
		},
		{
			name:                  "down at end of history clears input",
			inputValue:            "second message",
			inputHistory:          []string{"first message", "second message"},
			inputHistoryIndex:     1, // at last item
			msg:                   tea.KeyPressMsg{Code: tea.KeyDown},
			wantMode:              ModeInsert,
			wantInputValue:        "",
			wantMessageCount:      1,
			wantHistoryEmpty:      true,
			wantInputHistory:      []string{"first message", "second message"},
			wantInputHistoryIndex: 2, // past the end
		},
		{
			name:                  "enter with content adds to history",
			inputValue:            "new message",
			inputHistory:          []string{"old message"},
			inputHistoryIndex:     1,
			msg:                   tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"},
			wantMode:              ModeInsert,
			wantInputValue:        "",
			wantMessageCount:      2, // system + user
			wantHistoryEmpty:      false,
			wantInputHistory:      []string{"old message", "new message"},
			wantInputHistoryIndex: 2, // reset to end of history
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = ModeInsert
			model.ready = true
			model.input.SetValue(tt.inputValue)

			if tt.inputHistory != nil {
				model.inputHistory = tt.inputHistory
				model.inputHistoryIndex = tt.inputHistoryIndex
			}

			newModel, _ := model.Update(tt.msg)
			got := newModel.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if got.input.Value() != tt.wantInputValue {
				t.Errorf("input value = %q, want %q", got.input.Value(), tt.wantInputValue)
			}

			if len(got.messages) != tt.wantMessageCount {
				t.Errorf("message count = %d, want %d", len(got.messages), tt.wantMessageCount)
			}

			if tt.wantHistoryEmpty && got.history != "" {
				t.Errorf("history = %q, want empty", got.history)
			}

			if !tt.wantHistoryEmpty && got.history == "" {
				t.Error("history is empty, want non-empty")
			}

			if tt.wantInputHistory != nil {
				if len(got.inputHistory) != len(tt.wantInputHistory) {
					t.Errorf("inputHistory length = %d, want %d", len(got.inputHistory), len(tt.wantInputHistory))
				} else {
					for i, want := range tt.wantInputHistory {
						if got.inputHistory[i] != want {
							t.Errorf("inputHistory[%d] = %q, want %q", i, got.inputHistory[i], want)
						}
					}
				}

				if got.inputHistoryIndex != tt.wantInputHistoryIndex {
					t.Errorf("inputHistoryIndex = %d, want %d", got.inputHistoryIndex, tt.wantInputHistoryIndex)
				}
			}
		})
	}
}
