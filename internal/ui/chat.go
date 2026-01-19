package ui

import (
	"fmt"
	"strings"

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
	ModeInsert
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
	url       string
	model     string
}

// NewChatModel creates the chat model and initializes the text input.
func NewChatModel(url, model string) ChatModel {
	input := textinput.New()

	chatModel := ChatModel{
		input:    input,
		messages: []llm.ChatMessage{},
		history:  "",
		url:      url,
		model:    model,
	}

	return chatModel
}

// Init starts the cursor blink animation.
func (model ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and returns the updated model and optional command.
func (model ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height

		if !model.ready {
			model.viewport = viewport.New(viewport.WithWidth(model.width), viewport.WithHeight(model.height-3))

			model.ready = true
		}

		return model, nil

	case tea.KeyMsg:
		switch model.mode {
		case ModeNormal:
			switch msg.Key().Code {
			case ':':
				model.mode = ModeCommand
				model.cmdBuffer = ""

				return model, nil

			case 'i':
				model.mode = ModeInsert
				model.input.Focus()

				return model, textinput.Blink

			default:
				return model, nil
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

				return model, nil

			case tea.KeyEscape:
				model.mode = ModeNormal
				model.cmdBuffer = ""

				return model, nil

			default:
				model.cmdBuffer += msg.Key().Text

				return model, nil
			}

		case ModeInsert:
			switch msg.Key().Code {
			case tea.KeyEscape:
				model.mode = ModeNormal
				model.input.Blur()

				return model, nil

			case tea.KeyEnter:
				value := model.input.Value()

				if strings.TrimSpace(value) == "" {
					return model, nil
				}

				model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: value})
				model.history += fmt.Sprintf("You: %s\n", value)
				model.viewport.SetContent(model.history)

				model.input.SetValue("")

				return model, nil

			default:
				model.input, cmd = model.input.Update(msg)

				return model, cmd
			}
		}

	default:
		// Send messages for cursor blink
		if model.mode == ModeInsert {
			model.input, cmd = model.input.Update(msg)
		}

		return model, cmd
	}

	return model, nil
}

// View renders the current model state.
func (model ChatModel) View() tea.View {
	var view tea.View

	if !model.ready {
		view = tea.NewView("󱙝 initializing...")
		view.AltScreen = true

		return view
	}

	switch model.mode {
	case ModeNormal:
		view = tea.NewView(model.viewport.View() + "\n:[NORMAL]")
	case ModeCommand:
		view = tea.NewView(model.viewport.View() + "\n:" + model.cmdBuffer)
	case ModeInsert:
		view = tea.NewView(model.viewport.View() + "\n" + model.input.View())
	}

	view.AltScreen = true

	return view
}
