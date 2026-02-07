package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/theme"
)

func TestChatModel_HandleCommandMode(t *testing.T) {
	// Helper to create a temp file
	createTempFile := func(t *testing.T) string {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		err := os.WriteFile(path, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		return path
	}

	tests := []struct {
		name                 string
		inputValue           string
		setupFile            func(t *testing.T) string // returns file path to append to inputValue
		msg                  tea.Msg
		wantMode             Mode
		wantInputValue       string
		wantQuit             bool
		wantCmd              bool   // true if we expect any command (not just quit)
		wantChatHistoryMatch string // substring to check in chatHistory (empty = skip check)
		wantMessageCount     int    // 0 = skip check
		wantLastRole         llm.Role
	}{
		{
			name:           "q command quits",
			inputValue:     "q",
			msg:            tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:       ModeCommand,
			wantInputValue: "q",
			wantQuit:       true,
		},
		{
			name:           "invalid command returns to normal mode",
			inputValue:     "invalid",
			msg:            tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:       ModeNormal,
			wantInputValue: "",
			wantQuit:       false,
		},
		{
			name:           "escape returns to normal mode",
			inputValue:     "partial",
			msg:            tea.KeyPressMsg{Code: tea.KeyEscape},
			wantMode:       ModeNormal,
			wantInputValue: "",
			wantQuit:       false,
		},
		{
			name:           "typing appends to buffer",
			inputValue:     "q",
			msg:            tea.KeyPressMsg{Code: 'u', Text: "u"},
			wantMode:       ModeCommand,
			wantInputValue: "qu",
			wantQuit:       false,
			wantCmd:        true, // textinput returns cursor blink command
		},
		{
			name:                 "r without path shows error",
			inputValue:           "r",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error: no file path provided]", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with whitespace-only path shows error",
			inputValue:           "r   ",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error: no file path provided]", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with nonexistent file shows error",
			inputValue:           "r /nonexistent/file/path.txt",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error:", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with valid file loads content",
			inputValue:           "r ",
			setupFile:            createTempFile,
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: fmt.Sprintf("[%s loaded:", theme.GlyphInfo),
			wantMessageCount:     2,
			wantLastRole:         llm.RoleUser,
		},
		{
			name:       "r with directory shows error",
			inputValue: "r ",
			setupFile: func(t *testing.T) string {
				return t.TempDir()
			},
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error:", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:       "r with GIF file shows unsupported error",
			inputValue: "r ",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "test.gif")
				// GIF magic bytes
				gifBytes := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00}
				err := os.WriteFile(path, gifBytes, 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: "unsupported",
			wantMessageCount:     1,
		},
		{
			name:       "r with binary file shows unsupported error",
			inputValue: "r ",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "test.exe")
				// ELF magic bytes
				elfBytes := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00}
				err := os.WriteFile(path, elfBytes, 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantInputValue:       "",
			wantChatHistoryMatch: "unsupported",
			wantMessageCount:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel(t)
			model.mode = ModeCommand
			model.ready = true

			// Setup file and append path to inputValue if needed
			inputValue := tt.inputValue
			if tt.setupFile != nil {
				path := tt.setupFile(t)
				inputValue += path
			}
			model.cmdInput.SetValue(inputValue)

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if got.cmdInput.Value() != tt.wantInputValue {
				t.Errorf("cmdInput value = %q, want %q", got.cmdInput.Value(), tt.wantInputValue)
			}

			if tt.wantQuit && cmd == nil {
				t.Error("expected quit command, got nil")
			}

			if !tt.wantQuit && !tt.wantCmd && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}

			if tt.wantCmd && cmd == nil {
				t.Error("expected command, got nil")
			}

			if tt.wantChatHistoryMatch != "" && !strings.Contains(got.chatHistory, tt.wantChatHistoryMatch) {
				t.Errorf("chatHistory = %q, want to contain %q", got.chatHistory, tt.wantChatHistoryMatch)
			}

			if tt.wantMessageCount > 0 && len(got.messages) != tt.wantMessageCount {
				t.Errorf("messages count = %d, want %d", len(got.messages), tt.wantMessageCount)
			}

			if tt.wantLastRole != "" {
				lastMsg := got.messages[len(got.messages)-1]
				if lastMsg.Role != tt.wantLastRole {
					t.Errorf("last message role = %v, want %v", lastMsg.Role, tt.wantLastRole)
				}
			}
		})
	}
}
