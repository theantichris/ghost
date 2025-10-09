// Package tui implements the terminal user interface for Ghost's interactive
// chat mode using the BubbleTea framework. It handles user input, message
// display, viewport management, and LLM streaming integration.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
)

// streamingChunkMsg carries a single token from the LLM stream.
type streamingChunkMsg struct {
	content string
	sub     <-chan tea.Msg
}

// streamCompleteMsg signals that streaming is complete and carries the full accumulated response.
type streamCompleteMsg struct {
	content string
}

// streamErrorMsg carries error information when an LLM request fails.
type streamErrorMsg struct {
	err error
}

// Model represents the TUI application state for the chat interface.
// It implements the BubbleTea Model interface (Init, Update, View).
type Model struct {
	ctx         context.Context
	timeout     time.Duration
	logger      *log.Logger
	llmClient   llm.LLMClient
	chatHistory []llm.ChatMessage

	// UI state
	viewport viewport.Model // Chat message viewport
	messages []string       // Rendered messages for display
	input    string         // Current user input
	width    int            // Terminal width
	height   int            // Terminal height

	// Streaming state
	streaming  bool   // True if currently receiving a stream
	currentMsg string // Current message being streamed

	// Exit state
	exiting bool  // Indicates the model is exiting streaming
	err     error // Exit error
}

// NewModel creates a new TUI model initialized with the provided dependencies.
// The model is pre-configured with a system prompt and greeting instruction
// that will be sent to the LLM on initialization.
func NewModel(ctx context.Context, llmClient llm.LLMClient, timeout time.Duration, systemPrompt string, logger *log.Logger) Model {
	chatHistory := []llm.ChatMessage{
		{Role: llm.SystemRole, Content: systemPrompt},
		{Role: llm.SystemRole, Content: "Greet the user."},
	}

	viewport := viewport.New(80, 20)

	model := Model{
		ctx:         ctx,
		timeout:     timeout,
		llmClient:   llmClient,
		logger:      logger,
		viewport:    viewport,
		chatHistory: chatHistory,
	}

	return model
}

// Init initializes the TUI and returns a command to send the initial greeting.
// This is called once when the BubbleTea program starts.
func (model Model) Init() tea.Cmd {
	if len(model.chatHistory) > 0 {
		return model.sendChatRequest()
	}

	return nil
}

// Update handles all incoming messages and updates the model state accordingly.
// It processes terminal events (window resize, key presses) and custom messages
// (streaming chunks, completion, errors) from LLM requests.
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height

		// Save 3 lines for spacing, divider, and user input.
		viewportHeight := max(msg.Height-3, 1)

		model.viewport.Width = msg.Width
		model.viewport.Height = viewportHeight

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			model.input += string(msg.Runes)
		case tea.KeySpace:
			model.input += " "
		case tea.KeyBackspace:
			if len(model.input) > 0 {
				model.input = model.input[:len(model.input)-1]
			}
		case tea.KeyCtrlD, tea.KeyCtrlC:
			model.exiting = true
			return model, tea.Quit
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			var cmd tea.Cmd
			model.viewport, cmd = model.viewport.Update(msg)
			return model, cmd
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

				model.messages = append(model.messages, "You: "+input)

				model.input = ""

				model.viewport.SetContent(model.wordwrap())
				model.viewport.GotoBottom()

				return model, model.sendChatRequest()
			}
		}

	case streamingChunkMsg:
		model.streaming = true
		model.currentMsg += msg.content

		model.viewport.SetContent(model.wordwrap())
		model.viewport.GotoBottom()

		return model, waitForActivity(msg.sub)

	case streamCompleteMsg:
		model.streaming = false
		model.messages = append(model.messages, msg.content)

		model.chatHistory = append(model.chatHistory, llm.ChatMessage{
			Role:    llm.AssistantRole,
			Content: msg.content,
		})

		model.currentMsg = ""

		model.viewport.SetContent(model.wordwrap())
		model.viewport.GotoBottom()

	case streamErrorMsg:
		model.streaming = false
		model.err = msg.err
		model.currentMsg = ""
	}

	return model, nil
}

// View renders the TUI layout with the chat viewport, separator, and input field.
func (model Model) View() string {
	separator := strings.Repeat("─", model.width)

	view := model.viewport.View() + "\n" + separator + "\n" + model.input

	return view
}

// wordwrap formats all messages to fit the terminal width using lipgloss styling
// and returns the joined result as a single string for viewport rendering.
func (model Model) wordwrap() string {
	var wrapped []string

	for _, msg := range model.messages {
		wrappedMsg := lipgloss.NewStyle().Width(model.width).Render(msg)
		wrapped = append(wrapped, wrappedMsg)
	}

	if model.streaming && model.currentMsg != "" {
		wrappedCurrentMsg := lipgloss.NewStyle().Width(model.width).Render(model.currentMsg)
		wrapped = append(wrapped, wrappedCurrentMsg)
	}

	messages := strings.Join(wrapped, "\n")

	return messages
}

// sendChatRequest sends the current chat history to the LLM and accumulates the
// streamed response, returning streamCompleteMsg on success or streamErrorMsg on failure.
func (model Model) sendChatRequest() tea.Cmd {
	return func() tea.Msg {
		sub := make(chan tea.Msg)

		go func() {
			defer close(sub)

			if model.llmClient == nil {
				sub <- streamErrorMsg{err: ErrLLMClientInit}
				return
			}

			ctx, cancel := context.WithTimeout(model.ctx, model.timeout)
			defer cancel()

			var content strings.Builder

			err := model.llmClient.Chat(ctx, model.chatHistory, func(token string) {
				sub <- streamingChunkMsg{content: token, sub: sub}
				content.WriteString(token)
			})

			if err != nil {
				sub <- streamErrorMsg{err: fmt.Errorf("%w: %w", ErrLLMRequest, err)}

				return
			}

			sub <- streamCompleteMsg{content: content.String()}
		}()

		return <-sub
	}
}

// waitForActivity returns a command that waits for the next message from the subscription channel.
func waitForActivity(sub <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
