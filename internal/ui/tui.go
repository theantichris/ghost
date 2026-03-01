package ui

import (
	"context"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/storage"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/style"
)

// Mode represents the different modes the TUI can be in.
type Mode int

const (
	ModeNormal Mode = iota
	ModeCommand
	ModeInsert
	ModeThreadList
)

const inputHeight = 3
const statusHeight = 1

// TUIModel holds the TUI state.
type TUIModel struct {
	ctx               context.Context
	prompts           agent.Prompt
	logger            *log.Logger
	viewport          viewport.Model
	userInput         textarea.Model
	messages          []llm.ChatMessage
	chatHistory       string // Rendered conversation for display
	width             int
	height            int
	ready             bool // True if the viewport is initialized
	mode              Mode
	cmdInput          textinput.Model
	url               string
	chatLLM           string
	visionLLM         string
	responseCh        chan tea.Msg
	currentResponse   string // Buffer for the LLM's streaming response
	awaitingG         bool   // Used for gg command
	inputHistory      []string
	inputHistoryIndex int
	toolRegistry      tool.Registry
	store             *storage.Store
	threadID          string // ID of current conversation
	threadList        ThreadListModel
}

// NewTUIModel creates the chat model and initializes the text input.
func NewTUIModel(config ModelConfig) TUIModel {
	userInput := textarea.New()
	userInput.ShowLineNumbers = false
	userInput.SetHeight(inputHeight)
	userInput.SetStyles(textAreaStyles)

	cmdInput := textinput.New()
	cmdInput.Prompt = ":"
	cmdInput.Focus()

	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: config.Prompts.System},
	}

	chatModel := TUIModel{
		ctx:               config.Context,
		prompts:           config.Prompts,
		logger:            config.Logger,
		userInput:         userInput,
		cmdInput:          cmdInput,
		messages:          messages,
		chatHistory:       "",
		url:               config.URL,
		chatLLM:           config.ChatLLM,
		visionLLM:         config.VisionLLM,
		inputHistoryIndex: 0,
		toolRegistry:      config.Registry,
		store:             config.Store,
	}

	return chatModel
}

// Init starts the cursor blink animation.
func (model TUIModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and returns the updated model and optional command.
func (model TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return model.handleWindowSize(msg)

	case tea.KeyPressMsg:
		switch model.mode {
		case ModeNormal:
			return model.handleNormalMode(msg)

		case ModeCommand:
			return model.handleCommandMode(msg)

		case ModeInsert:
			return model.handleInsertMode(msg)

		case ModeThreadList:
			return model.handleThreadListMode(msg)
		}

	case LLMResponseMsg:
		return model.handleLLMResponseMsg(msg)

	case LLMDoneMsg:
		return model.handleLLMDoneMsg()

	case LLMErrorMsg:
		return model.handleLLMErrorMsg(msg)

	default:
		// Pass through to inputs
		var cmd tea.Cmd

		switch model.mode {
		case ModeInsert:
			model.userInput, cmd = model.userInput.Update(msg)

		case ModeCommand:
			model.cmdInput, cmd = model.cmdInput.Update(msg)

		case ModeThreadList:
			listModel, cmd := model.threadList.Update(msg)
			model.threadList = listModel.(ThreadListModel)

			return model, cmd
		}

		return model, cmd
	}

	return model, nil
}

// View renders the current model state.
func (model TUIModel) View() tea.View {
	var view tea.View

	if !model.ready {
		view = tea.NewView(baseBackground.Render(style.GlyphInfo + " initializing..."))
		view.AltScreen = true

		return view
	}

	switch model.mode {
	case ModeNormal:
		view = tea.NewView(
			lipgloss.JoinVertical(
				lipgloss.Left,
				viewportStyle.Width(model.panelWidth()).Render(model.viewport.View()),
				inputStyle.Width(model.panelWidth()).Render(model.userInput.View()),
				statusBarStyle.Width(model.panelWidth()).Render("[NOR]"),
			),
		)
	case ModeCommand:
		view = tea.NewView(
			lipgloss.JoinVertical(
				lipgloss.Left,
				viewportStyle.Width(model.panelWidth()).Render(model.viewport.View()),
				inputStyle.Width(model.panelWidth()).Render(model.userInput.View()),
				statusBarStyle.Width(model.panelWidth()).Render(model.cmdInput.View()),
			),
		)
	case ModeInsert:
		view = tea.NewView(
			lipgloss.JoinVertical(
				lipgloss.Left,
				viewportStyle.Width(model.panelWidth()).Render(model.viewport.View()),
				inputStyle.Width(model.panelWidth()).Render(model.userInput.View()),
				statusBarStyle.Width(model.panelWidth()).Render("[INS]"),
			),
		)
	case ModeThreadList:
		view = model.threadList.View()
	}

	view.AltScreen = true

	return view
}

func (model TUIModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	model.width = msg.Width
	model.height = msg.Height

	model.userInput.SetWidth(model.width - len(model.userInput.Prompt))

	if !model.ready {
		model.viewport = viewport.New(viewport.WithWidth(model.width), viewport.WithHeight(model.viewportHeight()))

		model.ready = true
	} else {
		model.viewport.SetWidth(msg.Width)
		model.viewport.SetHeight(model.viewportHeight())
	}

	return model, nil
}

func (model TUIModel) panelWidth() int {
	return model.width - horizontalChrome
}

func (model TUIModel) viewportHeight() int {
	return model.height - inputHeight - statusHeight - verticalChrome
}

// renderHistory returns the model history word wrapped to the width of the viewport.
func (model TUIModel) renderHistory() string {
	return lipgloss.NewStyle().Width(model.viewport.Width()).Render(model.chatHistory)
}
