package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

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

			if tt.wantHistoryEmpty && got.chatHistory != "" {
				t.Errorf("history = %q, want empty", got.chatHistory)
			}

			if !tt.wantHistoryEmpty && got.chatHistory == "" {
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
