package ui

import (
	"io"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/storage"
)

func TestNewThreadListModel(t *testing.T) {
	tests := []struct {
		name        string
		seedThreads int
		wantItems   int
		wantTitle   string
	}{
		{
			name:        "empty store returns empty list",
			seedThreads: 0,
			wantItems:   0,
			wantTitle:   "Threads",
		},
		{
			name:        "single thread returns one item",
			seedThreads: 1,
			wantItems:   1,
			wantTitle:   "Threads",
		},
		{
			name:        "multiple threads returns all items",
			seedThreads: 3,
			wantItems:   3,
			wantTitle:   "Threads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := storage.NewStore(t.TempDir())
			if err != nil {
				t.Fatalf("failed to create store: %v", err)
			}

			for i := range tt.seedThreads {
				_, err := store.CreateThread("thread " + string(rune('A'+i)))
				if err != nil {
					t.Fatalf("failed to create thread: %v", err)
				}
			}

			logger := log.New(io.Discard)

			model, err := NewThreadListModel(store, 80, 24, *logger)
			if err != nil {
				t.Fatalf("NewThreadListModel() error = %v", err)
			}

			if got := len(model.list.Items()); got != tt.wantItems {
				t.Errorf("item count = %d, want %d", got, tt.wantItems)
			}

			if got := model.list.Title; got != tt.wantTitle {
				t.Errorf("title = %q, want %q", got, tt.wantTitle)
			}
		})
	}
}

func TestThreadListModel_Update(t *testing.T) {
	tests := []struct {
		name    string
		msg     tea.Msg
		wantCmd bool
	}{
		{
			name:    "key message is delegated to list",
			msg:     tea.KeyPressMsg{Code: 'j'},
			wantCmd: false,
		},
		{
			name:    "window size message is delegated to list",
			msg:     tea.WindowSizeMsg{Width: 100, Height: 40},
			wantCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := storage.NewStore(t.TempDir())
			if err != nil {
				t.Fatalf("failed to create store: %v", err)
			}

			logger := log.New(io.Discard)

			model, err := NewThreadListModel(store, 80, 24, *logger)
			if err != nil {
				t.Fatalf("NewThreadListModel() error = %v", err)
			}

			newModel, cmd := model.Update(tt.msg)

			if newModel == nil {
				t.Fatal("Update() returned nil model")
			}

			if tt.wantCmd && cmd == nil {
				t.Error("expected command, got nil")
			}
		})
	}
}

func TestThreadListModel_View(t *testing.T) {
	store, err := storage.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	_, err = store.CreateThread("test thread")
	if err != nil {
		t.Fatalf("failed to create thread: %v", err)
	}

	logger := log.New(io.Discard)

	model, err := NewThreadListModel(store, 80, 24, *logger)
	if err != nil {
		t.Fatalf("NewThreadListModel() error = %v", err)
	}

	view := model.View()

	if view.Content == nil {
		t.Error("View() returned nil content")
	}
}
