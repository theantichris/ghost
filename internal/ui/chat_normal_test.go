package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestChatModel_HandleNormalMode(t *testing.T) {
	tests := []struct {
		name          string
		keys          []string
		wantMode      Mode
		wantAwaitingG bool
	}{
		{
			name:          "j scrolls down and stays in normal mode",
			keys:          []string{"j"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "k scrolls up and stays in normal mode",
			keys:          []string{"k"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "ctrl+d half page down and stays in normal mode",
			keys:          []string{"ctrl+d"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "ctrl+u half page up and stays in normal mode",
			keys:          []string{"ctrl+u"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "G goes to bottom and stays in normal mode",
			keys:          []string{"G"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "i enters insert mode",
			keys:          []string{"i"},
			wantMode:      ModeInsert,
			wantAwaitingG: false,
		},
		{
			name:          "colon enters command mode",
			keys:          []string{":"},
			wantMode:      ModeCommand,
			wantAwaitingG: false,
		},
		{
			name:          "single g sets awaitingG",
			keys:          []string{"g"},
			wantMode:      ModeNormal,
			wantAwaitingG: true,
		},
		{
			name:          "gg resets awaitingG",
			keys:          []string{"g", "g"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "g then other key resets awaitingG",
			keys:          []string{"g", "j"},
			wantMode:      ModeNormal,
			wantAwaitingG: false,
		},
		{
			name:          "g then i resets awaitingG and enters insert mode",
			keys:          []string{"g", "i"},
			wantMode:      ModeInsert,
			wantAwaitingG: false,
		},
		{
			name:          "g then G resets awaitingG",
			keys:          []string{"g", "G"},
			wantMode:      ModeNormal,
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

			got := result.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if got.awaitingG != tt.wantAwaitingG {
				t.Errorf("awaitingG = %v, want %v", got.awaitingG, tt.wantAwaitingG)
			}
		})
	}
}
