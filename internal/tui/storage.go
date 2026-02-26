package tui

import (
	"fmt"
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

func (model ChatModel) loadThread(threadID string) (ChatModel, error) {
	messages, err := model.store.GetMessages(threadID)
	if err != nil {
		model.logger.Error("failed to get messages", "thread_id", threadID, "error", err.Error())
		return model, err
	}

	var chatMessages []llm.ChatMessage
	var chatHistory strings.Builder
	for _, message := range messages {
		chatMessage := llm.ChatMessage{
			Role:      message.Role,
			Content:   message.Content,
			Images:    message.Images,
			ToolCalls: message.ToolCalls,
		}

		chatMessages = append(chatMessages, chatMessage)

		if message.Role == llm.RoleSystem || message.Role == llm.RoleTool {
			continue
		}

		label := "You"
		if message.Role == llm.RoleAssistant {
			label = "ghost"
		}

		history := fmt.Sprintf("%s: %s \n\n", label, message.Content)
		chatHistory.WriteString(history)
	}

	model.threadID = threadID
	model.messages = chatMessages
	model.chatHistory = chatHistory.String()

	return model, nil
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
