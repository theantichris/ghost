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

	config := ModelConfig{Context: context.Background(), URL: "http://localhost/11434/api", Model: "test-model", VisionModel: "test-vision-model", Registry: registry, Logger: logger}

	return NewChatModel(config)
}
