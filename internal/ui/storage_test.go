package ui

import (
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
