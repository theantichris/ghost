package ui

import tea "charm.land/bubbletea/v2"

func (model ChatModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Key().Code {
	case tea.KeyEnter:
		if model.cmdBuffer == "q" {
			model.logger.Info("disconnecting from ghost")

			return model, tea.Quit
		}

		// Invalid command, return to normal mode
		model.mode = ModeNormal
		model.cmdBuffer = ""

		return model, nil

	case tea.KeyEscape:
		model.mode = ModeNormal
		model.cmdBuffer = ""

		return model, nil

	default:
		model.cmdBuffer += msg.Key().Text

		return model, nil
	}
}
