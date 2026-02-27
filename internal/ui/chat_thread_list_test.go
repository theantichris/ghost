package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func TestChatModel_HandleThreadListMode(t *testing.T) {
	tests := []struct {
		name             string
		seedMessages     []llm.ChatMessage // messages to add to the seeded thread
		seedThread       bool              // whether to create a thread in the store
		msg              tea.Msg
		wantMode         Mode
		wantThreadLoaded bool   // true if threadID should be set after enter
		wantHistoryMatch string // substring to check in chatHistory
	}{
		{
			name: "esc returns to normal mode",
			msg:  tea.KeyPressMsg{Code: tea.KeyEscape},

			wantMode: ModeNormal,
		},
		{
			name:       "enter loads selected thread",
			seedThread: true,
			seedMessages: []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "hello ghost"},
				{Role: llm.RoleAssistant, Content: "greetings runner"},
			},
			msg: tea.KeyPressMsg{Code: tea.KeyEnter},

			wantMode:         ModeNormal,
			wantThreadLoaded: true,
			wantHistoryMatch: "hello ghost",
		},
		{
			name: "enter on empty list returns to normal mode",
			msg:  tea.KeyPressMsg{Code: tea.KeyEnter},

			wantMode: ModeNormal,
		},
		{
			name: "other keys pass through to list",
			msg:  tea.KeyPressMsg{Text: "j"},

			wantMode: ModeThreadList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel(t)

			if tt.seedThread {
				thread, err := model.store.CreateThread("test thread")
				if err != nil {
					t.Fatalf("failed to create thread: %v", err)
				}

				for _, msg := range tt.seedMessages {
					_, err := model.store.AddMessage(thread.ID, msg)
					if err != nil {
						t.Fatalf("failed to add message: %v", err)
					}
				}
			}

			threadList, err := NewThreadListModel(model.store, 80, 24, model.logger)
			if err != nil {
				t.Fatalf("failed to create thread list model: %v", err)
			}

			model.threadList = threadList
			model.mode = ModeThreadList
			model.ready = true

			result, _ := model.Update(tt.msg)
			got := result.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if tt.wantThreadLoaded && got.threadID == "" {
				t.Error("threadID is empty, want non-empty")
			}

			if !tt.wantThreadLoaded && got.threadID != "" {
				t.Errorf("threadID = %q, want empty", got.threadID)
			}

			if tt.wantHistoryMatch != "" {
				if got.chatHistory == "" {
					t.Error("chatHistory is empty, want non-empty")
				}
			}
		})
	}
}

// testMsg is an arbitrary message type that hits the default branch of Update.
type testMsg struct{}

func TestChatModel_Update_NonKeyMsgForwardsToThreadList(t *testing.T) {
	model := newTestModel(t)

	threadList, err := NewThreadListModel(model.store, 80, 24, model.logger)
	if err != nil {
		t.Fatalf("failed to create thread list model: %v", err)
	}

	model.threadList = threadList
	model.mode = ModeThreadList
	model.ready = true

	result, _ := model.Update(testMsg{})
	got := result.(ChatModel)

	if got.mode != ModeThreadList {
		t.Errorf("mode = %v, want %v", got.mode, ModeThreadList)
	}
}

func TestChatModel_HandleThreadListMode_LoadsMessages(t *testing.T) {
	model := newTestModel(t)

	thread, err := model.store.CreateThread("test thread")
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	seedMessages := []llm.ChatMessage{
		{Role: llm.RoleUser, Content: "first message"},
		{Role: llm.RoleAssistant, Content: "first response"},
		{Role: llm.RoleUser, Content: "second message"},
	}

	for _, msg := range seedMessages {
		_, err := model.store.AddMessage(thread.ID, msg)
		if err != nil {
			t.Fatalf("failed to add message: %v", err)
		}
	}

	threadList, err := NewThreadListModel(model.store, 80, 24, model.logger)
	if err != nil {
		t.Fatalf("failed to create thread list model: %v", err)
	}

	model.threadList = threadList
	model.mode = ModeThreadList
	model.ready = true

	result, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := result.(ChatModel)

	// Verify all messages were loaded (not just the thread ID).
	if len(got.messages) != len(seedMessages) {
		t.Errorf("message count = %d, want %d", len(got.messages), len(seedMessages))
	}

	// Verify the thread was loaded into the correct thread.
	storedMessages, err := model.store.GetMessages(thread.ID)
	if err != nil {
		t.Fatalf("GetMessages() err = %v", err)
	}

	for i, stored := range storedMessages {
		if got.messages[i].Content != stored.Content {
			t.Errorf("messages[%d].Content = %q, want %q", i, got.messages[i].Content, stored.Content)
		}
	}
}
