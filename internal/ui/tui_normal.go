package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var normalKeyMap = keyMap{
	command: key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "command mode"),
	),
	insert: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "insert mode"),
	),
	down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/down", "scroll down"),
	),
	up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/up", "scroll down"),
	),
	pageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "page down"),
	),
	pageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "page up"),
	),
	goToTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("gg", "go to top"),
	),
	goToBottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to bottom"),
	),
}

func (model TUIModel) handleNormalMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	wasAwaitingG := model.awaitingG
	model.awaitingG = false

	switch {
	case key.Matches(msg, normalKeyMap.command):
		model.mode = ModeCommand
		model.cmdInput.Reset()
		model.cmdInput.Focus()

		return model, textinput.Blink

	case key.Matches(msg, normalKeyMap.insert):
		model.mode = ModeInsert
		model.userInput.Focus()

		return model, textinput.Blink

	case key.Matches(msg, normalKeyMap.down):
		model.viewport.ScrollDown(1)

	case key.Matches(msg, normalKeyMap.up):
		model.viewport.ScrollUp(1)

	case key.Matches(msg, normalKeyMap.pageDown):
		model.viewport.HalfPageDown()

	case key.Matches(msg, normalKeyMap.pageUp):
		model.viewport.HalfPageUp()

	case key.Matches(msg, normalKeyMap.goToTop):
		if wasAwaitingG {
			model.viewport.GotoTop()
		} else {
			model.awaitingG = true
		}

	case key.Matches(msg, normalKeyMap.goToBottom):
		model.viewport.GotoBottom()
	}

	return model, nil
}
