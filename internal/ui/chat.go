package ui

import (
	"context"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/theme"
)

// Mode represents the different modes the TUI can be in.
type Mode int

const inputHeight = 3

const (
	ModeNormal Mode = iota
	ModeCommand
	ModeInsert
)

// LLMResponseMsg carries a chunk of the LLM response.
type LLMResponseMsg string

// LLMErrorMsg signals an error from the LLM.
type LLMErrorMsg struct {
	Err error
}

// ChatModel holds the TUI state.
type ChatModel struct {
	ctx               context.Context
	logger            *log.Logger
	viewport          viewport.Model
	input             textarea.Model
	messages          []llm.ChatMessage
	chatHistory       string // Rendered conversation for display
	width             int
	height            int
	ready             bool // True if the viewport is initialized
	mode              Mode
	cmdInput          textinput.Model
	url               string
	model             string
	visionModel       string
	responseCh        chan tea.Msg
	currentResponse   string
	awaitingG         bool
	inputHistory      []string
	inputHistoryIndex int
	toolRegistry      tool.Registry
}

// NewChatModel creates the chat model and initializes the text input.
func NewChatModel(config ModelConfig) ChatModel {
	input := textarea.New()
	input.ShowLineNumbers = false
	input.SetHeight(2)

	cmdInput := textinput.New()
	cmdInput.Prompt = ":"
	cmdInput.Focus()

	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: config.System},
	}

	chatModel := ChatModel{
		ctx:               config.Context,
		logger:            config.Logger,
		input:             input,
		cmdInput:          cmdInput,
		messages:          messages,
		chatHistory:       "",
		url:               config.URL,
		model:             config.Model,
		visionModel:       config.VisionModel,
		inputHistoryIndex: 0,
		toolRegistry:      config.Registry,
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
		return model.handleWindowSize(msg)

	case tea.KeyMsg:
		switch model.mode {
		case ModeNormal:
			return model.handleNormalMode(msg)

		case ModeCommand:
			return model.handleCommandMode(msg)

		case ModeInsert:
			return model.handleInsertMode(msg)
		}

	case LLMResponseMsg:
		return model.handleLLMResponseMsg(msg)

	case LLMDoneMsg:
		return model.handleLLMDoneMsg()

	case LLMErrorMsg:
		return model.handleLLMErrorMsg(msg)

	default:
		// Pass through to textinputs
		var cmd tea.Cmd

		if model.mode == ModeInsert {
			model.input, cmd = model.input.Update(msg)
		}

		if model.mode == ModeCommand {
			model.cmdInput, cmd = model.cmdInput.Update(msg)
		}

		return model, cmd
	}

	return model, nil
}

// View renders the current model state.
func (model ChatModel) View() tea.View {
	var view tea.View

	if !model.ready {
		view = tea.NewView(theme.GlyphInfo + " initializing...")
		view.AltScreen = true

		return view
	}

	switch model.mode {
	case ModeNormal:
		view = tea.NewView(model.viewport.View() + "\n" + model.input.View() + "\n[NOR]")
	case ModeCommand:
		view = tea.NewView(model.viewport.View() + "\n" + model.input.View() + "\n" + model.cmdInput.View())
	case ModeInsert:
		view = tea.NewView(model.viewport.View() + "\n" + model.input.View() + "\n[INS]")
	}

	view.AltScreen = true

	return view
}

func (model ChatModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	model.width = msg.Width
	model.height = msg.Height

	model.input.SetWidth(model.width - len(model.input.Prompt))

	if !model.ready {
		model.viewport = viewport.New(viewport.WithWidth(model.width), viewport.WithHeight(model.height-inputHeight))

		model.ready = true
	} else {
		model.viewport.SetWidth(msg.Width)
		model.viewport.SetHeight(msg.Height - inputHeight)
	}

	return model, nil
}

// renderHistory returns the model history word wrapped to the width of the viewport.
func (model ChatModel) renderHistory() string {
	return lipgloss.NewStyle().Width(model.viewport.Width()).Render(model.chatHistory)
}
