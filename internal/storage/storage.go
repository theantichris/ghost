package storage

import (
	"errors"
	"time"

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
