package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func (model ChatModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Key().Code {
	case tea.KeyEnter:
		parts := strings.SplitN(model.cmdBuffer, " ", 2)
		cmd := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		switch cmd {
		case "q":
			model.logger.Info("disconnecting from ghost")

			return model, tea.Quit

		case "r":
			if arg == "" {
				model.chatHistory += "\n[󱙜 error: no file path provided]\n"
				model.viewport.SetContent(model.renderHistory())
				model.mode = ModeNormal
				model.cmdBuffer = ""

				return model, nil
			}

			content, err := agent.ReadFileForContext(arg)
			if err != nil {
				model.logger.Error("file read failed", "path", arg, "error", err)
				model.chatHistory += fmt.Sprintf("\n[󱙜 error: %s]\n", err.Error())
				model.viewport.SetContent(model.renderHistory())
				model.mode = ModeNormal
				model.cmdBuffer = ""

				return model, nil
			}

			model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: content})
			model.logger.Info("file loaded into context", "path", arg)

			model.chatHistory += fmt.Sprintf("\n[󱙝 loaded: %s]\n", arg)
			model.viewport.SetContent(model.renderHistory())
		}

		model.mode = ModeNormal
		model.cmdBuffer = ""

	case tea.KeyEscape:
		model.mode = ModeNormal
		model.cmdBuffer = ""

	default:
		model.cmdBuffer += msg.Key().Text
	}

	return model, nil
}
