package ui

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/internal/llm"
)

// ChatModel holds the TUI state.
type ChatModel struct {
	viewport viewport.Model
	input    textinput.Model
	messages []llm.ChatMessage
	history  string // Rendered conversation for display
	width    int
	height   int
	ready    bool // True if the viewport is initialized
}

// NewChatModel creates the chat model and initializes the text input.
func NewChatModel() ChatModel {
	input := textinput.New()
	input.Placeholder = "enter message..."
	input.Focus()

	chatModel := ChatModel{
		input:    input,
		messages: []llm.ChatMessage{},
		history:  "",
	}

	return chatModel
}

// Init starts the cursor blink animation.
func (model ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and returns the updated model and optional command.
func (model ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height

		if !model.ready {
			model.viewport = viewport.New(viewport.WithWidth(model.width), viewport.WithHeight(model.height-3))

			model.ready = true
		}
	}

	return model, nil
}

// View renders the current model state.
func (model ChatModel) View() tea.View {
	if !model.ready {
		return tea.NewView("󱙝 initializing")
	}

	return tea.NewView(model.viewport.View() + "\n" + model.input.View())
}
