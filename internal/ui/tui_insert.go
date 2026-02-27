package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/llm"
)

var insertKeyMap = keyMap{
	esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "normal mode"),
	),
	newline: key.NewBinding(
		key.WithKeys("shift+enter", "ctrl+j"),
		key.WithHelp("shift+enter/ctrl+j", "add newline"),
	),
	up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "input history back"),
	),
	down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "input history forward"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send message"),
	),
}

func (model TUIModel) handleInsertMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, insertKeyMap.esc):
		model.mode = ModeNormal
		model.userInput.Blur()

	case key.Matches(msg, insertKeyMap.newline):
		value := model.userInput.Value() + "\n"
		model.userInput.SetValue(value)

		return model, nil

	case key.Matches(msg, insertKeyMap.up):
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex--
		if model.inputHistoryIndex < 0 {
			model.inputHistoryIndex = 0
		}

		model.userInput.SetValue(model.inputHistory[model.inputHistoryIndex])

	case key.Matches(msg, insertKeyMap.down):
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

	case key.Matches(msg, insertKeyMap.enter):
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
