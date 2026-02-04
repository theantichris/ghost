package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/theme"
)

func (model ChatModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Key().Code {
	case tea.KeyEnter:
		parts := strings.SplitN(model.cmdInput.Value(), " ", 2)
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
			return model.readFile(arg)
		}

		// Resets mode for invalid commands.
		model.mode = ModeNormal
		model.cmdInput.Reset()

	case tea.KeyEscape:
		model.mode = ModeNormal
		model.cmdInput.Reset()

	default:
		var cmd tea.Cmd
		model.cmdInput, cmd = model.cmdInput.Update(msg)

		return model, cmd
	}

	return model, nil
}

func (model ChatModel) readFile(arg string) (tea.Model, tea.Cmd) {
	if arg == "" {
		model.chatHistory += fmt.Sprintf("\n[%s error: no file path provided]\n", theme.GlyphError)
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	fileType, err := agent.DetectFileType(arg)
	if err != nil {
		model.logger.Error("failed to validate file", "error", err.Error(), "path", arg)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", theme.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	switch fileType {
	case agent.FileTypeDir:
		model.chatHistory += fmt.Sprintf("\n[%s error: file is directory]\n", theme.GlyphError)
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

func (model ChatModel) analyzeImage(path string) (tea.Model, tea.Cmd) {
	content, err := agent.AnalyseImages(model.ctx, model.url, model.visionModel, []string{path}, model.logger)
	if err != nil {
		model.logger.Error("image read failed", "path", path, "error", err)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", theme.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}

	model.messages = append(model.messages, content...)
	model.logger.Info("image loaded into context", "path", path)

	model.chatHistory += fmt.Sprintf("\n[%s loaded image: %s]\n", theme.GlyphInfo, path)
	model.viewport.SetContent(model.renderHistory())

	model.mode = ModeNormal
	model.cmdInput.Reset()

	return model, nil
}

func (model ChatModel) readTextFile(path string) (tea.Model, tea.Cmd) {
	content, err := agent.ReadTextFile(path)
	if err != nil {
		model.logger.Error("file read failed", "path", path, "error", err)
		model.chatHistory += fmt.Sprintf("\n[%s error: %s]\n", theme.GlyphError, err.Error())
		model.viewport.SetContent(model.renderHistory())
		model.mode = ModeNormal
		model.cmdInput.Reset()

		return model, nil
	}
	model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: content})
	model.logger.Info("loaded file", "path", path)

	model.chatHistory += fmt.Sprintf("\n[%s loaded: %s]\n", theme.GlyphInfo, path)
	model.viewport.SetContent(model.renderHistory())

	model.mode = ModeNormal
	model.cmdInput.Reset()

	return model, nil
}
