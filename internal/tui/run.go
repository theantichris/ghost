package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theantichris/ghost/internal/llm"
)

// Run starts the BubbleTea TUI program with the provided model.
// It runs in alternate screen mode and returns any error that occurs.
func Run(model Model, systemPrompt string, greetingPrompt string) error {
	model.chatHistory = []llm.ChatMessage{
		{Role: llm.SystemRole, Content: systemPrompt},
		{Role: llm.SystemRole, Content: greetingPrompt},
	}

	_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()

	return err
}
