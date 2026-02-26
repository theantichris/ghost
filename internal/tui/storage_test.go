package tui

import (
	"strings"
	"testing"

	"github.com/theantichris/ghost/v3/internal/llm"
)

func TestChatModel_SaveMessage(t *testing.T) {
	tests := []struct {
		name             string
		messages         []llm.ChatMessage
		wantThreadTitle  string
		wantMessageCount int
	}{
		{
			name: "first message creates thread and saves",
			messages: []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "hello ghost"},
			},
			wantThreadTitle:  "hello ghost",
			wantMessageCount: 1,
		},
		{
			name: "multiple messages saved to same thread",
			messages: []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "hello ghost"},
				{Role: llm.RoleAssistant, Content: "greetings runner"},
				{Role: llm.RoleUser, Content: "tell me more"},
			},
			wantThreadTitle:  "hello ghost",
			wantMessageCount: 3,
		},
		{
			name: "long first message truncates thread title at word boundary",
			messages: []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "this is a very long message that should be truncated when used as a thread title"},
			},
			wantThreadTitle:  "this is a very long message that should be",
			wantMessageCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel(t)

			for _, msg := range tt.messages {
				model = model.saveMessage(msg)
			}

			if model.threadID == "" {
				t.Fatal("saveMessage() threadID is empty, want non-empty")
			}

			thread, err := model.store.GetThread(model.threadID)
			if err != nil {
				t.Fatalf("saveMessage() GetThread() returned error: %v", err)
			}

			if thread.Title != tt.wantThreadTitle {
				t.Errorf("saveMessage() thread title = %q, want %q", thread.Title, tt.wantThreadTitle)
			}

			messages, err := model.store.GetMessages(model.threadID)
			if err != nil {
				t.Fatalf("saveMessage() GetMessages() returned error: %v", err)
			}

			if len(messages) != tt.wantMessageCount {
				t.Errorf("saveMessage() message count = %d, want %d", len(messages), tt.wantMessageCount)
			}
		})
	}
}

func TestChatModel_CreateThread(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantTitle string
	}{
		{
			name:      "short content used as full title",
			content:   "hello",
			wantTitle: "hello",
		},
		{
			name:      "content truncated at word boundary",
			content:   "this is a very long message that should be truncated when used as a thread title",
			wantTitle: "this is a very long message that should be",
		},
		{
			name:      "single word exceeding limit produces empty title",
			content:   "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			wantTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel(t)

			thread, err := model.createThread(tt.content)
			if err != nil {
				t.Fatalf("createThread() returned error: %v", err)
			}

			if thread.Title != tt.wantTitle {
				t.Errorf("createThread() title = %q, want %q", thread.Title, tt.wantTitle)
			}
		})
	}
}

func TestChatModel_LoadThread(t *testing.T) {
	tests := []struct {
		name               string
		seedMessages       []llm.ChatMessage
		threadID           string // empty means use seeded thread
		wantErr            bool
		wantMessageCount   int
		wantHistoryContain []string
		wantHistoryExclude []string
	}{
		{
			name: "loads user and assistant messages",
			seedMessages: []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "hello ghost"},
				{Role: llm.RoleAssistant, Content: "greetings runner"},
			},
			wantMessageCount:   2,
			wantHistoryContain: []string{"You: hello ghost", "ghost: greetings runner"},
		},
		{
			name: "excludes system and tool messages from history",
			seedMessages: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "you are a cyberpunk AI"},
				{Role: llm.RoleUser, Content: "hello ghost"},
				{Role: llm.RoleTool, Content: "tool output here"},
			},
			wantMessageCount:   3,
			wantHistoryContain: []string{"You: hello ghost"},
			wantHistoryExclude: []string{"you are a cyberpunk AI", "tool output here"},
		},
		{
			name:     "invalid thread ID returns error",
			threadID: "nonexistent-id",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel(t)

			threadID := tt.threadID

			if threadID == "" {
				thread, err := model.store.CreateThread("test thread")
				if err != nil {
					t.Fatalf("failed to create thread: %v", err)
				}

				threadID = thread.ID

				for _, msg := range tt.seedMessages {
					_, err := model.store.AddMessage(threadID, msg)
					if err != nil {
						t.Fatalf("failed to add message: %v", err)
					}
				}
			}

			model, err := model.loadThread(threadID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("loadThread() err = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("loadThread() err = %v, want nil", err)
			}

			if model.threadID != threadID {
				t.Errorf("loadThread() threadID = %q, want %q", model.threadID, threadID)
			}

			if len(model.messages) != tt.wantMessageCount {
				t.Errorf("loadThread() message count = %d, want %d", len(model.messages), tt.wantMessageCount)
			}

			for _, want := range tt.wantHistoryContain {
				if !strings.Contains(model.chatHistory, want) {
					t.Errorf("loadThread() chatHistory missing %q", want)
				}
			}

			for _, exclude := range tt.wantHistoryExclude {
				if strings.Contains(model.chatHistory, exclude) {
					t.Errorf("loadThread() chatHistory should not contain %q", exclude)
				}
			}
		})
	}
}
