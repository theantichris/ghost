package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/style"
)

var threadListKeyMap = keyMap{
	esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "normal model"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select thread"),
	),
}

func (model TUIModel) handleThreadListMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, threadListKeyMap.esc):
		model.mode = ModeNormal

		return model, nil

	case key.Matches(msg, threadListKeyMap.enter):
		selectedThread, ok := model.threadList.list.SelectedItem().(threadItem)
		if ok {
			var err error
			model, err = model.loadThread(selectedThread.thread.ID)
			if err != nil {
				model.logger.Error("error loading thread", "thread_id", selectedThread.thread.ID, "error", err.Error())
				model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", style.GlyphError, err.Error())
			}
		}

		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	// Pass through to the list model update
	listModel, cmd := model.threadList.Update(msg)
	model.threadList = listModel.(ThreadListModel)

	return model, cmd
}
