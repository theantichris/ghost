package ui

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/storage"
	"github.com/theantichris/ghost/v3/internal/tool"
)

// ModelConfig holds the configuration options for the UI models.
type ModelConfig struct {
	Context     context.Context
	Logger      *log.Logger
	URL         string
	Model       string
	VisionModel string
	Format      string
	System      string
	Messages    []llm.ChatMessage
	Images      []string
	Registry    tool.Registry
	Store       *storage.Store
}
