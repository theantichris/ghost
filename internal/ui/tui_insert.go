package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func (model TUIModel) handleInsertMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		model.mode = ModeNormal
		model.userInput.Blur()

	case "shift+enter", "ctrl+j":
		value := model.userInput.Value() + "\n"
		model.userInput.SetValue(value)

		return model, nil

	case "up":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex--
		if model.inputHistoryIndex < 0 {
			model.inputHistoryIndex = 0
		}

		model.userInput.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "down":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex++
		if model.inputHistoryIndex >= len(model.inputHistory) {
			model.inputHistoryIndex = len(model.inputHistory)
			model.userInput.SetValue("")

			return model, nil
		}

		model.userInput.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "enter":
		value := model.userInput.Value()

		if strings.TrimSpace(value) == "" {
			return model, nil
		}

		model.inputHistory = append(model.inputHistory, value)
		model.inputHistoryIndex = len(model.inputHistory)

		model.userInput.SetValue("")
		userMsg := llm.ChatMessage{Role: llm.RoleUser, Content: value}
		model.messages = append(model.messages, userMsg)
		model = model.saveMessage(userMsg)
		model.chatHistory += fmt.Sprintf("You: %s\n\nghost: ", value)
		model.viewport.SetContent(model.renderHistory())

		return model, model.startLLMStream()

	default:
		model.userInput, cmd = model.userInput.Update(msg)

		return model, cmd
	}

	return model, nil
}
