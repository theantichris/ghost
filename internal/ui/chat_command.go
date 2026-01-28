package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (model ChatModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Key().Code {
	case tea.KeyEnter:
		parts := strings.SplitN(model.cmdBuffer, " ", 2)
		cmd := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		switch cmd {
		case "q":
			model.logger.Info("disconnecting from ghost")

			return model, tea.Quit

		case "r":
			_ = arg
		}

		// Invalid command, return to normal mode
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
