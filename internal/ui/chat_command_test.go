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
		cmdBuffer            string
		setupFile            func(t *testing.T) string // returns file path to append to cmdBuffer
		msg                  tea.Msg
		wantMode             Mode
		wantCmdBuffer        string
		wantQuit             bool
		wantChatHistoryMatch string // substring to check in chatHistory (empty = skip check)
		wantMessageCount     int    // 0 = skip check
		wantLastRole         llm.Role
	}{
		{
			name:          "q command quits",
			cmdBuffer:     "q",
			msg:           tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:      ModeCommand,
			wantCmdBuffer: "q",
			wantQuit:      true,
		},
		{
			name:          "invalid command returns to normal mode",
			cmdBuffer:     "invalid",
			msg:           tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:      ModeNormal,
			wantCmdBuffer: "",
			wantQuit:      false,
		},
		{
			name:          "escape returns to normal mode",
			cmdBuffer:     "partial",
			msg:           tea.KeyPressMsg{Code: tea.KeyEscape},
			wantMode:      ModeNormal,
			wantCmdBuffer: "",
			wantQuit:      false,
		},
		{
			name:          "typing appends to buffer",
			cmdBuffer:     "q",
			msg:           tea.KeyPressMsg{Code: 'u', Text: "u"},
			wantMode:      ModeCommand,
			wantCmdBuffer: "qu",
			wantQuit:      false,
		},
		{
			name:                 "r without path shows error",
			cmdBuffer:            "r",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantCmdBuffer:        "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error: no file path provided]", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with whitespace-only path shows error",
			cmdBuffer:            "r   ",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantCmdBuffer:        "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error: no file path provided]", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with nonexistent file shows error",
			cmdBuffer:            "r /nonexistent/file/path.txt",
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantCmdBuffer:        "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error:", theme.GlyphError),
			wantMessageCount:     1,
		},
		{
			name:                 "r with valid file loads content",
			cmdBuffer:            "r ",
			setupFile:            createTempFile,
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantCmdBuffer:        "",
			wantChatHistoryMatch: fmt.Sprintf("[%s loaded:", theme.GlyphInfo),
			wantMessageCount:     2,
			wantLastRole:         llm.RoleUser,
		},
		{
			name:      "r with directory shows error",
			cmdBuffer: "r ",
			setupFile: func(t *testing.T) string {
				return t.TempDir()
			},
			msg:                  tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:             ModeNormal,
			wantCmdBuffer:        "",
			wantChatHistoryMatch: fmt.Sprintf("[%s error:", theme.GlyphError),
			wantMessageCount:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = ModeCommand
			model.ready = true

			// Setup file and append path to cmdBuffer if needed
			cmdBuffer := tt.cmdBuffer
			if tt.setupFile != nil {
				path := tt.setupFile(t)
				cmdBuffer += path
			}
			model.cmdBuffer = cmdBuffer

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if got.cmdBuffer != tt.wantCmdBuffer {
				t.Errorf("cmdBuffer = %q, want %q", got.cmdBuffer, tt.wantCmdBuffer)
			}

			if tt.wantQuit && cmd == nil {
				t.Error("expected quit command, got nil")
			}

			if !tt.wantQuit && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
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
