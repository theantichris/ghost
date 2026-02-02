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
	cmdBuffer         string
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
func NewChatModel(ctx context.Context, url, model, visionModel, system string, registry tool.Registry, logger *log.Logger) ChatModel {
	input := textarea.New()
	input.ShowLineNumbers = false
	input.SetHeight(2)

	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: system},
	}

	chatModel := ChatModel{
		ctx:               ctx,
		logger:            logger,
		input:             input,
		messages:          messages,
		chatHistory:       "",
		url:               url,
		model:             model,
		visionModel:       visionModel,
		inputHistoryIndex: 0,
		toolRegistry:      registry,
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
		// Pass through to textinput
		var cmd tea.Cmd

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
		view = tea.NewView(theme.GlyphInfo + " initializing...")
		view.AltScreen = true

		return view
	}

	switch model.mode {
	case ModeNormal:
		view = tea.NewView(model.viewport.View() + "\n[NORMAL]\n" + model.input.View())
	case ModeCommand:
		view = tea.NewView(model.viewport.View() + "\n:" + model.cmdBuffer + "\n" + model.input.View())
	case ModeInsert:
		view = tea.NewView(model.viewport.View() + "\n[INSERT]\n" + model.input.View())
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
