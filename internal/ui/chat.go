package ui

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/internal/llm"
)

// Mode represents the different modes the TUI can be in.
type Mode int

const (
	ModeNormal Mode = iota
	ModeCommand
)

// ChatModel holds the TUI state.
type ChatModel struct {
	viewport  viewport.Model
	input     textinput.Model
	messages  []llm.ChatMessage
	history   string // Rendered conversation for display
	width     int
	height    int
	ready     bool // True if the viewport is initialized
	mode      Mode
	cmdBuffer string
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

	case tea.KeyMsg:
		switch model.mode {
		case ModeNormal:
			if msg.Key().Code == ':' {
				model.mode = ModeCommand
				model.cmdBuffer = ""
			}

		case ModeCommand:
			switch msg.Key().Code {
			case tea.KeyEnter:
				if model.cmdBuffer == "q" {
					return model, tea.Quit
				}

				// Invalid command, return to normal mode
				model.mode = ModeNormal
				model.cmdBuffer = ""

			case tea.KeyEscape:
				model.mode = ModeNormal
				model.cmdBuffer = ""

			default:
				model.cmdBuffer += msg.Key().Text
			}
		}
	}

	return model, nil
}

// View renders the current model state.
func (model ChatModel) View() tea.View {
	var view tea.View

	if !model.ready {
		view = tea.NewView("󱙝 initializing")
		view.AltScreen = true

		return view
	}

	if model.mode == ModeCommand {
		view = tea.NewView(model.viewport.View() + "\n:" + model.cmdBuffer)
	} else {
		view = tea.NewView(model.viewport.View() + "\n" + model.input.View())
	}

	view.AltScreen = true

	return view
}
