package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func (model ChatModel) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		model.mode = ModeNormal
		model.input.Blur()

	case "shift+enter", "ctrl+j":
		value := model.input.Value() + "\n"
		model.input.SetValue(value)

		return model, nil

	case "up":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex--
		if model.inputHistoryIndex < 0 {
			model.inputHistoryIndex = 0
		}

		model.input.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "down":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex++
		if model.inputHistoryIndex >= len(model.inputHistory) {
			model.inputHistoryIndex = len(model.inputHistory)
			model.input.SetValue("")

			return model, nil
		}

		model.input.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "enter":
		value := model.input.Value()

		if strings.TrimSpace(value) == "" {
			return model, nil
		}

		model.inputHistory = append(model.inputHistory, value)
		model.inputHistoryIndex = len(model.inputHistory)

		model.input.SetValue("")
		model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: value})
		model.chatHistory += fmt.Sprintf("You: %s\n\nghost: ", value)
		model.viewport.SetContent(model.renderHistory())

		return model, model.startLLMStream()

	default:
		model.input, cmd = model.input.Update(msg)

		return model, cmd
	}

	return model, nil
}
