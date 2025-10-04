package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	// logger      *log.Logger
	// llmClient   *llm.LLMClient
	// chatHistory []llm.ChatMessage

	// UI state
	// messages []string // Rendered messages for display
	input  string // Current user input
	width  int    // Terminal width
	height int    // Terminal height

	// Streaming state
	// streaming  bool   // True if currently receiving a stream
	// currentMsg string // Current message being streamed

	// Exit state
	// exiting bool
	// err     error
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
	case tea.KeyMsg:
		model.input = string(msg.Runes)

	}

	return model, nil
}

func (model Model) View() string {
	return ""
}
