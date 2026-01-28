package ui

import (
	"errors"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/theme"
)

func TestChatModel_HandleLLMMessages(t *testing.T) {
	tests := []struct {
		name                string
		currentResponse     string
		chatHistory         string
		msg                 tea.Msg
		wantChatHistory     string
		wantCurrentResponse string
		wantMessageCount    int
		wantLastRole        llm.Role
		wantLastContent     string
		wantCmd             bool
	}{
		{
			name:                "response msg appends to history and current response",
			currentResponse:     "",
			chatHistory:         "",
			msg:                 LLMResponseMsg("hello "),
			wantChatHistory:     "hello ",
			wantCurrentResponse: "hello ",
			wantMessageCount:    1,
			wantCmd:             true,
		},
		{
			name:                "done msg finalizes response and adds assistant message",
			currentResponse:     "test response",
			chatHistory:         "You: hi\n\nghost: test response",
			msg:                 LLMDoneMsg{},
			wantChatHistory:     "You: hi\n\nghost: test response\n\n",
			wantCurrentResponse: "",
			wantMessageCount:    2,
			wantLastRole:        llm.RoleAssistant,
			wantLastContent:     "test response",
			wantCmd:             false,
		},
		{
			name:                "error msg adds error to history",
			currentResponse:     "",
			chatHistory:         "",
			msg:                 LLMErrorMsg{Err: errors.New("test error")},
			wantChatHistory:     fmt.Sprintf("\n[%s error: test error]\n", theme.GlyphInfo),
			wantCurrentResponse: "",
			wantMessageCount:    1,
			wantCmd:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.ready = true
			model.currentResponse = tt.currentResponse
			model.chatHistory = tt.chatHistory
			model.responseCh = make(chan tea.Msg)

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(ChatModel)

			if got.chatHistory != tt.wantChatHistory {
				t.Errorf("chatHistory = %q, want %q", got.chatHistory, tt.wantChatHistory)
			}

			if got.currentResponse != tt.wantCurrentResponse {
				t.Errorf("currentResponse = %q, want %q", got.currentResponse, tt.wantCurrentResponse)
			}

			if len(got.messages) != tt.wantMessageCount {
				t.Errorf("messages length = %d, want %d", len(got.messages), tt.wantMessageCount)
			}

			if tt.wantLastRole != "" {
				lastMsg := got.messages[len(got.messages)-1]
				if lastMsg.Role != tt.wantLastRole {
					t.Errorf("last message role = %v, want %v", lastMsg.Role, tt.wantLastRole)
				}
				if lastMsg.Content != tt.wantLastContent {
					t.Errorf("last message content = %q, want %q", lastMsg.Content, tt.wantLastContent)
				}
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
