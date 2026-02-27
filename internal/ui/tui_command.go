package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/style"
)

var commandKeyMap = keyMap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit command"),
	),
	new: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new chat"),
	),
	quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	readFile: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "read file"),
	),
	threadList: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "open thread list"),
	),
	esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "normal mode"),
	),
}

func (model TUIModel) handleCommandMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, commandKeyMap.enter):
		parts := strings.SplitN(model.cmdInput.Value(), " ", 2)
		cmd := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		switch {
		case matchesCommand(cmd, commandKeyMap.new):
			return model.newChat()

		case matchesCommand(cmd, commandKeyMap.quit):
			model.logger.Info("disconnecting from ghost")

			return model, tea.Quit

		case matchesCommand(cmd, commandKeyMap.readFile):
			return model.readFile(arg)

		case matchesCommand(cmd, commandKeyMap.threadList):
			return model.createThreadList()
		}

		// Resets mode for invalid commands.
		model.mode = ModeNormal
		model.cmdInput.Reset()

	case key.Matches(msg, commandKeyMap.esc):
		model.mode = ModeNormal
		model.cmdInput.Reset()

	default:
		var cmd tea.Cmd
		model.cmdInput, cmd = model.cmdInput.Update(msg)

		return model, cmd
	}

	return model, nil
}

func (model TUIModel) createThreadList() (tea.Model, tea.Cmd) {
	threadList, err := NewThreadListModel(model.store, model.width, model.height, model.logger)
	if err != nil {
		model.logger.Error("error creating thread list", "error", err)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", style.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	model.mode = ModeThreadList
	model.threadList = threadList
	model.cmdInput.Reset()

	return model, nil
}

func (model TUIModel) readFile(arg string) (tea.Model, tea.Cmd) {
	if arg == "" {
		model.chatHistory += fmt.Sprintf("\n[%s error: no file path provided]\n", style.GlyphError)
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	fileType, err := agent.DetectFileType(arg)
	if err != nil {
		model.logger.Error("failed to validate file", "error", err.Error(), "path", arg)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", style.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	switch fileType {
	case agent.FileTypeDir:
		model.chatHistory += fmt.Sprintf("\n[%s error: file is directory]\n", style.GlyphError)
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

	case agent.FileTypeImage:
		return model.analyzeImage(arg)

	case agent.FileTypeText:
		return model.readTextFile(arg)
	}

	return model, nil
}

func (model TUIModel) analyzeImage(path string) (tea.Model, tea.Cmd) {
	content, err := agent.AnalyseImages(model.ctx, model.url, model.visionLLM, model.prompts, []string{path}, model.logger)
	if err != nil {
		model.logger.Error("image read failed", "path", path, "error", err)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", style.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	model.messages = append(model.messages, content...)
	model.logger.Info("image loaded into context", "path", path)

	model.chatHistory += fmt.Sprintf("\n[%s loaded image: %s]\n", style.GlyphInfo, path)
	model.viewport.SetContent(model.renderHistory())

	model.mode = ModeNormal
	model.cmdInput.Reset()

	return model, nil
}

func (model TUIModel) readTextFile(path string) (tea.Model, tea.Cmd) {
	content, err := agent.ReadTextFile(path)
	if err != nil {
		model.logger.Error("file read failed", "path", path, "error", err)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", style.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}
	model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: content})
	model.logger.Info("loaded file", "path", path)

	model.chatHistory += fmt.Sprintf("\n[%s loaded: %s]\n", style.GlyphInfo, path)
	model.viewport.SetContent(model.renderHistory())

	model.mode = ModeNormal
	model.cmdInput.Reset()

	return model, nil
}

func (model TUIModel) newChat() (tea.Model, tea.Cmd) {
	model.messages = []llm.ChatMessage{{Role: llm.RoleSystem, Content: model.prompts.System}}
	model.chatHistory = ""
	model.threadID = ""
	model.viewport.SetContent("")
	model.cmdInput.Reset()
	model.mode = ModeNormal

	return model, nil
}
