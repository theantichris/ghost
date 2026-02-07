package ui

import (
	"strings"

	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/storage"
)

func (model ChatModel) saveMessage(chatMsg llm.ChatMessage) ChatModel {
	if model.threadID == "" {
		thread, err := model.createThread(chatMsg.Content)
		if err != nil {
			return model
		}

		model.threadID = thread.ID
	}

	_, err := model.store.AddMessage(model.threadID, chatMsg)
	if err != nil {
		model.logger.Error("failed to add message to thread", "thread_id", model.threadID, "error", err)
	}

	return model
}

func (model ChatModel) createThread(content string) (*storage.Thread, error) {
	words := strings.Fields(content)
	title := ""

	for _, word := range words {
		if len(title)+len(word)+1 > 50 {
			break
		}

		if title != "" {
			title += " "
		}

		title += word
	}

	thread, err := model.store.CreateThread(title)
	if err != nil {
		model.logger.Error("failed to create new thread", "error", err)
	}

	return thread, err
}
