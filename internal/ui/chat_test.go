package ui

import (
	"context"
	"io"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/storage"
	"github.com/theantichris/ghost/v3/internal/tool"
)

func newTestModel(t *testing.T) ChatModel {
	t.Helper()

	logger := log.New(io.Discard)
	registry := tool.NewRegistry("", 0, logger)

	store, err := storage.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	config := ModelConfig{Context: context.Background(), URL: "http://localhost/11434/api", Model: "test-model", VisionModel: "test-vision-model", Registry: registry, Logger: logger, Store: store}

	return NewChatModel(config)
}
