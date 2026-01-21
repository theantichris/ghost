package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/theantichris/ghost/internal/llm"
)

var (
	ErrThreadNotFound = errors.New("thread not found in memory banks")
	ErrStorageAccess  = errors.New("failed to access data storage")
	ErrCorruptedData  = errors.New("corrupted data detected in storage")
)

// Thread represents a conversation thread.
type Thread struct {
	ID        string    `json:"id"`    // UUID
	Title     string    `json:"title"` // User facing name
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message wraps llm.ChatMessage with storage metadata.
type Message struct {
	ID        string         `json:"id"`        // UUID
	ThreadID  string         `json:"thread_id"` // Foreign key to Thread
	Role      llm.Role       `json:"role"`
	Content   string         `json:"content"`
	Images    []string       `json:"images,omitempty"`
	ToolCalls []llm.ToolCall `json:"tool_calls,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// Conversation is the wrapper that bundles a thread with its messages.
type Conversation struct {
	Thread   Thread    `json:"thread"`
	Messages []Message `json:"messages"`
}

// Store manages all file operations.
type Store struct {
	threadsDir string
	mu         sync.RWMutex
}

// NewStore creates the threads directory in the base directory if it doesn't
// exist then creates and returns a new store.
func NewStore(baseDir string) (*Store, error) {
	threadsDir := filepath.Join(baseDir, "threads")

	err := os.MkdirAll(threadsDir, 0750)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	store := Store{
		threadsDir: threadsDir,
	}

	return &store, nil
}

// threadPath returns a path to a thread JSON files.
func (store *Store) threadPath(id string) string {
	return filepath.Join(store.threadsDir, id+".json")
}

// readConversation retrieves and returns a Conversation from a JSON file.
// Assume the caller has acquired the lock.
func (store *Store) readConversation(threadId string) (Conversation, error) {
	bytes, err := os.ReadFile(store.threadPath(threadId))
	if err != nil {
		if os.IsNotExist(err) {
			return Conversation{}, fmt.Errorf("%w: %w", ErrThreadNotFound, err)
		}

		return Conversation{}, fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	var conversation Conversation
	err = json.Unmarshal(bytes, &conversation)
	if err != nil {
		return Conversation{}, fmt.Errorf("%w: %w", ErrCorruptedData, err)
	}

	return conversation, nil
}

// writeConversation writes a Conversation to a JSON file.
// Assumes the caller has acquired the lock.
func (store *Store) writeConversation(conversation Conversation) error {
	bytes, err := json.MarshalIndent(conversation, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCorruptedData, err)
	}

	err = os.WriteFile(store.threadPath(conversation.Thread.ID), bytes, 0640)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	return nil
}

// CreateThread creates a new Thread and Conversation and writes them to storage.
func (store *Store) CreateThread(title string) (*Thread, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	id := uuid.New().String()

	thread := Thread{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	conversation := Conversation{
		Thread:   thread,
		Messages: []Message{},
	}

	err := store.writeConversation(conversation)
	if err != nil {
		return nil, err
	}

	return &thread, nil
}

// GetThread retrieves a Conversation from storage and returns the Thread.
func (store *Store) GetThread(id string) (*Thread, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	conversation, err := store.readConversation(id)
	if err != nil {
		return nil, err
	}

	return &conversation.Thread, nil
}

// UpdateThread updates the Thread in the Conversation and writes it to storage.
func (store *Store) UpdateThread(thread *Thread) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	conversation, err := store.readConversation(thread.ID)
	if err != nil {
		return err
	}

	thread.UpdatedAt = time.Now()

	conversation.Thread = *thread

	err = store.writeConversation(conversation)
	if err != nil {
		return err
	}

	return nil
}

// DeleteThread deletes a Thread from storage.
func (store *Store) DeleteThread(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	path := store.threadPath(id)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %w", ErrThreadNotFound, err)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	return nil
}

// ListThreads returns a slice of all threads in storage sort.
// The slice is sorted with most recent thread first.
func (store *Store) ListThreads() ([]Thread, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	dirEntries, err := os.ReadDir(store.threadsDir)
	if err != nil {
		return []Thread{}, fmt.Errorf("%w: %w", ErrStorageAccess, err)
	}

	var threads []Thread

	for _, entry := range dirEntries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")

		conversation, err := store.readConversation(id)
		if err != nil {
			return []Thread{}, err
		}

		threads = append(threads, conversation.Thread)
	}

	sort.Slice(threads, func(x, y int) bool {
		return threads[x].UpdatedAt.After(threads[y].UpdatedAt)
	})

	return threads, nil
}

// AddMessage adds a new Message to a Conversation.
func (store *Store) AddMessage(threadID string, chatMsg llm.ChatMessage) (*Message, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	conversation, err := store.readConversation(threadID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	message := Message{
		ID:        uuid.New().String(),
		ThreadID:  threadID,
		Role:      chatMsg.Role,
		Content:   chatMsg.Content,
		Images:    chatMsg.Images,
		ToolCalls: chatMsg.ToolCalls,
		CreatedAt: now,
	}

	conversation.Messages = append(conversation.Messages, message)
	conversation.Thread.UpdatedAt = now

	err = store.writeConversation(conversation)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// GetMessages returns all Messages from a Conversation.
func (store *Store) GetMessages(threadID string) ([]Message, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	conversation, err := store.readConversation(threadID)
	if err != nil {
		return []Message{}, err
	}

	return conversation.Messages, nil
}
