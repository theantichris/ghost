package tui

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/storage"
	"github.com/theantichris/ghost/v3/internal/tool"
)

// ModelConfig holds the configuration options for the UI models.
type ModelConfig struct {
	Context   context.Context
	Logger    *log.Logger
	URL       string
	ChatLLM   string
	VisionLLM string
	Format    string
	Prompts   agent.Prompt
	Messages  []llm.ChatMessage
	Images    []string
	Registry  tool.Registry
	Store     *storage.Store
}
