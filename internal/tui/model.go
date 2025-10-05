package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
)

type Model struct {
	logger *log.Logger
	// llmClient   *llm.LLMClient
	chatHistory []llm.ChatMessage

	// UI state
	messages []string // Rendered messages for display
	input    string   // Current user input
	width    int      // Terminal width
	height   int      // Terminal height

	// Streaming state
	// streaming  bool   // True if currently receiving a stream
	// currentMsg string // Current message being streamed

	// Exit state
	exiting bool
	// err     error
}

func NewModel(systemPrompt string, logger *log.Logger) Model {
	chatHistory := []llm.ChatMessage{
		{Role: llm.SystemRole, Content: systemPrompt},
		{Role: llm.SystemRole, Content: "Greet the user."},
	}

	model := Model{
		logger:      logger,
		chatHistory: chatHistory,
	}

	return model
}

func (model Model) Init() tea.Cmd {
	if len(model.chatHistory) > 0 {
		return model.sendChatRequest
	}

	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			model.input += string(msg.Runes)
		case tea.KeyBackspace:
			if len(model.input) > 0 {
				model.input = model.input[:len(model.input)-1]
			}
		case tea.KeyCtrlD, tea.KeyCtrlC:
			model.exiting = true
			return model, tea.Quit
		case tea.KeyEnter:
			if model.input != "" {
				input := model.input

				if input == "/bye" || input == "/exit" {
					model.exiting = true
					input = "Goodbye!"
				}

				model.chatHistory = append(model.chatHistory, llm.ChatMessage{
					Role:    llm.UserRole,
					Content: input,
				})

				model.input = ""
			}
		}
	}

	return model, nil
}

func (model Model) View() string {
	chatArea := strings.Join(model.messages, "\n")
	separator := strings.Repeat("─", model.width)

	view := chatArea + "\n" + separator + "\n" + model.input

	return view
}

func (model Model) sendChatRequest() tea.Msg {
	return nil
}
