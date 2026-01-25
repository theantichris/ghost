package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theantichris/ghost/v3/internal/llm"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	return store
}

func TestNewStore(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			name: "creates store and threads directory",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
		},
		{
			name: "fails with invalid path",
			setupDir: func(t *testing.T) string {
				return "/nonexistent/path/that/cannot/be/created"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.setupDir(t)

			store, err := NewStore(baseDir)

			if tt.wantErr {
				if err == nil {
					t.Error("NewStore() err = nil, want error")
				}

				if !errors.Is(err, ErrStorageAccess) {
					t.Errorf("NewStore() err = %v, want %v", err, ErrStorageAccess)
				}

				return
			}

			if err != nil {
				t.Fatalf("NewStore() err = %v, want nil", err)
			}

			threadsDir := filepath.Join(baseDir, "threads")

			info, err := os.Stat(threadsDir)
			if err != nil {
				t.Fatalf("threads directory not created: %v", err)
			}

			if !info.IsDir() {
				t.Error("threads path is not a directory")
			}

			if store.threadsDir != threadsDir {
				t.Errorf("store.threadsDir = %s, want %s", store.threadsDir, threadsDir)
			}
		})
	}
}

func TestCreateThread(t *testing.T) {
	store := setupTestStore(t)

	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "creates thread with title",
			title: "Test Thread",
		},
		{
			name:  "creates thread with empty title",
			title: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thread, err := store.CreateThread(tt.title)
			if err != nil {
				t.Fatalf("CreateThread() err = %v, want nil", err)
			}

			if thread.ID == "" {
				t.Error("CreateThread() returned empty ID")
			}

			if thread.Title != tt.title {
				t.Errorf("CreateThread() Title = %s, want %s", thread.Title, tt.title)
			}

			if thread.CreatedAt.IsZero() {
				t.Error("CreateThread() CreatedAt is zero")
			}

			if thread.UpdatedAt.IsZero() {
				t.Error("CreateThread() UpdatedAt is zero")
			}

			// Verify file was created
			path := store.threadPath(thread.ID)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("CreateThread() did not create file on disk")
			}
		})
	}
}

func TestGetThread(t *testing.T) {
	store := setupTestStore(t)

	created, err := store.CreateThread("Test Thread")
	if err != nil {
		t.Fatalf("CreateThread() err = %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		err     error
	}{
		{
			name: "returns thread",
			id:   created.ID,
		},
		{
			name:    "returns error for nonexistent thread",
			id:      "nonexistent-id",
			wantErr: true,
			err:     ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetThread(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetThread() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("GetThread() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("GetThread() err = %v, want nil", err)
			}

			if got.ID != created.ID {
				t.Errorf("GetThread() ID = %s, want %s", got.ID, created.ID)
			}

			if got.Title != created.Title {
				t.Errorf("GetThread() Title = %s, want %s", got.Title, created.Title)
			}
		})
	}
}

func TestUpdateThread(t *testing.T) {
	store := setupTestStore(t)

	created, err := store.CreateThread("Original Title")
	if err != nil {
		t.Fatalf("CreateThread() err = %v", err)
	}

	originalUpdatedAt := created.UpdatedAt

	// Small delay to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	nonexistent := &Thread{ID: "nonexistent-id", Title: "Does Not Exist"}

	tests := []struct {
		name    string
		thread  *Thread
		wantErr bool
		err     error
	}{
		{
			name:   "updates thread",
			thread: &Thread{ID: created.ID, Title: "Updated Title", CreatedAt: created.CreatedAt},
		},
		{
			name:    "returns error for nonexistent thread",
			thread:  nonexistent,
			wantErr: true,
			err:     ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdateThread(tt.thread)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateThread() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("UpdateThread() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("UpdateThread() err = %v, want nil", err)
			}

			got, err := store.GetThread(tt.thread.ID)
			if err != nil {
				t.Fatalf("GetThread() err = %v", err)
			}

			if got.Title != "Updated Title" {
				t.Errorf("UpdateThread() Title = %s, want %s", got.Title, "Updated Title")
			}

			if !got.UpdatedAt.After(originalUpdatedAt) {
				t.Error("UpdateThread() did not update UpdatedAt timestamp")
			}
		})
	}
}

func TestDeleteThread(t *testing.T) {
	store := setupTestStore(t)

	created, err := store.CreateThread("To Be Deleted")
	if err != nil {
		t.Fatalf("CreateThread() err = %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		err     error
	}{
		{
			name: "deletes thread",
			id:   created.ID,
		},
		{
			name:    "returns error for nonexistent thread",
			id:      "nonexistent-id",
			wantErr: true,
			err:     ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteThread(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteThread() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("DeleteThread() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("DeleteThread() err = %v, want nil", err)
			}

			path := store.threadPath(tt.id)
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Error("DeleteThread() did not remove file from disk")
			}
		})
	}
}

func TestListThreads(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Store) error
		wantLen   int
		wantFirst string
		wantLast  string
	}{
		{
			name:    "returns empty list",
			setup:   func(s *Store) error { return nil },
			wantLen: 0,
		},
		{
			name: "returns threads sorted by most recent",
			setup: func(s *Store) error {
				_, err := s.CreateThread("First")
				if err != nil {
					return err
				}

				time.Sleep(10 * time.Millisecond)

				_, err = s.CreateThread("Second")
				if err != nil {
					return err
				}

				time.Sleep(10 * time.Millisecond)

				_, err = s.CreateThread("Third")
				return err
			},
			wantLen:   3,
			wantFirst: "Third",
			wantLast:  "First",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupTestStore(t)

			if err := tt.setup(store); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			threads, err := store.ListThreads()
			if err != nil {
				t.Fatalf("ListThreads() err = %v, want nil", err)
			}

			if len(threads) != tt.wantLen {
				t.Fatalf("ListThreads() len = %d, want %d", len(threads), tt.wantLen)
			}

			if tt.wantLen > 0 {
				if threads[0].Title != tt.wantFirst {
					t.Errorf("ListThreads() first thread = %s, want %s", threads[0].Title, tt.wantFirst)
				}

				if threads[len(threads)-1].Title != tt.wantLast {
					t.Errorf("ListThreads() last thread = %s, want %s", threads[len(threads)-1].Title, tt.wantLast)
				}
			}
		})
	}
}

func TestAddMessage(t *testing.T) {
	store := setupTestStore(t)

	thread, err := store.CreateThread("Test Thread")
	if err != nil {
		t.Fatalf("CreateThread() err = %v", err)
	}

	originalUpdatedAt := thread.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name     string
		threadID string
		chatMsg  llm.ChatMessage
		wantErr  bool
		err      error
	}{
		{
			name:     "adds message",
			threadID: thread.ID,
			chatMsg:  llm.ChatMessage{Role: llm.RoleUser, Content: "Hello, Ghost!"},
		},
		{
			name:     "adds message with images and tool calls",
			threadID: thread.ID,
			chatMsg: llm.ChatMessage{
				Role:    llm.RoleAssistant,
				Content: "Here's the analysis",
				Images:  []string{"base64encodedimage"},
				ToolCalls: []llm.ToolCall{
					{
						Function: struct {
							Name      string          `json:"name"`
							Arguments json.RawMessage `json:"arguments"`
						}{
							Name:      "search",
							Arguments: json.RawMessage(`{"query": "test"}`),
						},
					},
				},
			},
		},
		{
			name:     "returns error for nonexistent thread",
			threadID: "nonexistent-id",
			chatMsg:  llm.ChatMessage{Role: llm.RoleUser, Content: "Hello"},
			wantErr:  true,
			err:      ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := store.AddMessage(tt.threadID, tt.chatMsg)

			if tt.wantErr {
				if err == nil {
					t.Error("AddMessage() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("AddMessage() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("AddMessage() err = %v, want nil", err)
			}

			if msg.ID == "" {
				t.Error("AddMessage() returned empty ID")
			}

			if msg.ThreadID != tt.threadID {
				t.Errorf("AddMessage() ThreadID = %s, want %s", msg.ThreadID, tt.threadID)
			}

			if msg.Role != tt.chatMsg.Role {
				t.Errorf("AddMessage() Role = %s, want %s", msg.Role, tt.chatMsg.Role)
			}

			if msg.Content != tt.chatMsg.Content {
				t.Errorf("AddMessage() Content = %s, want %s", msg.Content, tt.chatMsg.Content)
			}

			if msg.CreatedAt.IsZero() {
				t.Error("AddMessage() CreatedAt is zero")
			}

			if len(tt.chatMsg.Images) > 0 && len(msg.Images) != len(tt.chatMsg.Images) {
				t.Errorf("AddMessage() Images len = %d, want %d", len(msg.Images), len(tt.chatMsg.Images))
			}

			if len(tt.chatMsg.ToolCalls) > 0 && len(msg.ToolCalls) != len(tt.chatMsg.ToolCalls) {
				t.Errorf("AddMessage() ToolCalls len = %d, want %d", len(msg.ToolCalls), len(tt.chatMsg.ToolCalls))
			}

			// Verify thread timestamp updated
			updated, err := store.GetThread(tt.threadID)
			if err != nil {
				t.Fatalf("GetThread() err = %v", err)
			}

			if !updated.UpdatedAt.After(originalUpdatedAt) {
				t.Error("AddMessage() did not update thread's UpdatedAt timestamp")
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	store := setupTestStore(t)

	thread, err := store.CreateThread("Test Thread")
	if err != nil {
		t.Fatalf("CreateThread() err = %v", err)
	}

	// Add messages for the success case
	_, err = store.AddMessage(thread.ID, llm.ChatMessage{Role: llm.RoleUser, Content: "Hello"})
	if err != nil {
		t.Fatalf("AddMessage() err = %v", err)
	}

	_, err = store.AddMessage(thread.ID, llm.ChatMessage{Role: llm.RoleAssistant, Content: "Hi there!"})
	if err != nil {
		t.Fatalf("AddMessage() err = %v", err)
	}

	tests := []struct {
		name         string
		threadID     string
		wantLen      int
		wantContents []string
		wantErr      bool
		err          error
	}{
		{
			name:         "returns messages in order",
			threadID:     thread.ID,
			wantLen:      2,
			wantContents: []string{"Hello", "Hi there!"},
		},
		{
			name:     "returns error for nonexistent thread",
			threadID: "nonexistent-id",
			wantErr:  true,
			err:      ErrThreadNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := store.GetMessages(tt.threadID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetMessages() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("GetMessages() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("GetMessages() err = %v, want nil", err)
			}

			if len(messages) != tt.wantLen {
				t.Fatalf("GetMessages() len = %d, want %d", len(messages), tt.wantLen)
			}

			for i, content := range tt.wantContents {
				if messages[i].Content != content {
					t.Errorf("GetMessages()[%d].Content = %s, want %s", i, messages[i].Content, content)
				}
			}
		})
	}
}
