package ui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
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

// LLMDoneMsg signals the LLM request is complete.
type LLMDoneMsg struct{}

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
	history           string // Rendered conversation for display
	width             int
	height            int
	ready             bool // True if the viewport is initialized
	mode              Mode
	cmdBuffer         string
	url               string
	model             string
	responseCh        chan tea.Msg
	currentResponse   string
	awaitingG         bool
	inputHistory      []string
	inputHistoryIndex int
}

// NewChatModel creates the chat model and initializes the text input.
func NewChatModel(ctx context.Context, url, model, system string, logger *log.Logger) ChatModel {
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
		history:           "",
		url:               url,
		model:             model,
		inputHistoryIndex: 0,
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
		view = tea.NewView("󱙝 initializing...")
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

// listenForChunk returns a command that waits for the next chunk from the channel.
func listenForChunk(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return LLMDoneMsg{}
		}

		return msg
	}
}

// startLLMStream starts the LLM call in a go routine.
// It returns the first listenForChunk command to start receiving.
func (model *ChatModel) startLLMStream() tea.Cmd {
	model.logger.Debug("transmitting to neural network", "model", model.model, "messages", len(model.messages))

	model.responseCh = make(chan tea.Msg)

	go func() {
		ch := model.responseCh

		_, err := llm.StreamChat(
			model.ctx,
			model.url,
			model.model,
			model.messages,
			nil,
			func(chunk string) {
				ch <- LLMResponseMsg(chunk)
			},
		)

		if err != nil {
			ch <- LLMErrorMsg{Err: err}
		}

		close(ch)
	}()

	return listenForChunk(model.responseCh)
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

func (model ChatModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	wasAwaitingG := model.awaitingG
	model.awaitingG = false

	switch msg.String() {
	case ":":
		model.mode = ModeCommand
		model.cmdBuffer = ""

	case "i":
		model.mode = ModeInsert
		model.input.Focus()

		return model, textinput.Blink

	case "j":
		model.viewport.ScrollDown(1)

	case "k":
		model.viewport.ScrollUp(1)

	case "ctrl+d":
		model.viewport.HalfPageDown()

	case "ctrl+u":
		model.viewport.HalfPageUp()

	case "g":
		if wasAwaitingG {
			model.viewport.GotoTop()
		} else {
			model.awaitingG = true
		}

	case "G":
		model.viewport.GotoBottom()
	}

	return model, nil
}

func (model ChatModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Key().Code {
	case tea.KeyEnter:
		if model.cmdBuffer == "q" {
			model.logger.Info("disconnecting from ghost")

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
}

func (model ChatModel) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		model.mode = ModeNormal
		model.input.Blur()

	case "shift+enter", "ctrl+j":
		value := model.input.Value() + "\n"
		model.input.SetValue(value)

		return model, nil

	case "up":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex--
		if model.inputHistoryIndex < 0 {
			model.inputHistoryIndex = 0
		}

		model.input.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "down":
		if len(model.inputHistory) == 0 {
			return model, nil
		}

		model.inputHistoryIndex++
		if model.inputHistoryIndex > len(model.inputHistory)-1 {
			model.inputHistoryIndex = len(model.inputHistory) - 1
		}

		model.input.SetValue(model.inputHistory[model.inputHistoryIndex])

	case "enter":
		value := model.input.Value()

		if strings.TrimSpace(value) == "" {
			return model, nil
		}

		model.inputHistory = append(model.inputHistory, value)
		model.inputHistoryIndex = len(model.inputHistory)
		model.logger.Debug("updated input history", "length", len(model.inputHistory), "index", model.inputHistoryIndex)

		model.input.SetValue("")
		model.messages = append(model.messages, llm.ChatMessage{Role: llm.RoleUser, Content: value})
		model.history += fmt.Sprintf("You: %s\n\nghost: ", value)
		model.viewport.SetContent(model.renderHistory())

		return model, model.startLLMStream()

	default:
		model.input, cmd = model.input.Update(msg)

		return model, cmd
	}

	return model, nil
}

func (model ChatModel) handleLLMResponseMsg(msg LLMResponseMsg) (tea.Model, tea.Cmd) {
	model.history += string(msg)
	model.currentResponse += string(msg)
	model.viewport.SetContent(model.renderHistory())
	model.viewport.GotoBottom()

	return model, listenForChunk(model.responseCh)
}

func (model ChatModel) handleLLMDoneMsg() (tea.Model, tea.Cmd) {
	model.logger.Debug("transmission complete", "response_length", len(model.currentResponse))

	model.history += "\n\n"
	model.viewport.SetContent(model.renderHistory())
	model.messages = append(model.messages, llm.ChatMessage{
		Role:    llm.RoleAssistant,
		Content: model.currentResponse,
	})

	model.currentResponse = ""

	return model, nil
}

func (model ChatModel) handleLLMErrorMsg(msg LLMErrorMsg) (tea.Model, tea.Cmd) {
	model.logger.Error("neural link disrupted", "error", msg.Err)

	model.history += fmt.Sprintf("\n[󱙝 error: %v]\n", msg.Err)
	model.viewport.SetContent(model.renderHistory())

	return model, nil
}

// renderHistory returns the model history word wrapped to the width of the viewport.
func (model ChatModel) renderHistory() string {
	return lipgloss.NewStyle().Width(model.viewport.Width()).Render(model.history)
}
