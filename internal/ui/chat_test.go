package ui

import (
	"context"
	"io"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/tool"
)

func newTestModel() ChatModel {
	logger := log.New(io.Discard)
	registry := tool.NewRegistry("", 0, logger)

	return NewChatModel(context.Background(), "http://localhost:11434/api", "test-model", "test-vision-model", "test system", registry, logger)
}
