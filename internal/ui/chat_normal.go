package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func (model ChatModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	wasAwaitingG := model.awaitingG
	model.awaitingG = false

	switch msg.String() {
	case ":":
		model.mode = ModeCommand
		model.cmdBuffer = ""

	case "i":
		model.mode = ModeInsert
		model.input.Focus()

		return model, textinput.Blink

	case "j":
		model.viewport.ScrollDown(1)

	case "k":
		model.viewport.ScrollUp(1)

	case "ctrl+d":
		model.viewport.HalfPageDown()

	case "ctrl+u":
		model.viewport.HalfPageUp()

	case "g":
		if wasAwaitingG {
			model.viewport.GotoTop()
		} else {
			model.awaitingG = true
		}

	case "G":
		model.viewport.GotoBottom()
	}

	return model, nil
}
