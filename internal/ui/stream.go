package ui

import (
	"context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/theme"
)

var quitKeys = key.NewBinding(
	key.WithKeys("ctrl+c"),
)

// StreamChunkMsg represents a chunk of text received from the LLM.
type StreamChunkMsg string

// StreamErrorMsg signals an error occurred during streaming.
type StreamErrorMsg struct {
	Err error
}

// StreamModel handles the UI for streaming LLM responses.
type StreamModel struct {
	ctx          context.Context
	logger       *log.Logger // Logger for error visibility.
	width        int         // Terminal width
	content      string      // Accumulated response content.
	messages     []llm.ChatMessage
	url          string
	model        string
	visionModel  string
	images       []string
	toolRegistry tool.Registry
	done         bool          // Whether streaming has finished.
	Err          error         // Error if streaming failed.
	spinner      spinner.Model // Animated spinner.
	format       string        // Format for output.
	responseCh   chan tea.Msg
}

// NewStreamModel creates and returns StreamModel.
func NewStreamModel(config ModelConfig) StreamModel {
	s := spinner.New()
	s.Spinner = spinner.Ellipsis
	s.Style = theme.FgAccent0

	return StreamModel{
		ctx:          config.Context,
		logger:       config.Logger,
		width:        80,
		messages:     config.Messages,
		url:          config.URL,
		model:        config.Model,
		visionModel:  config.VisionModel,
		images:       config.Images,
		toolRegistry: config.Registry,
		content:      "",
		done:         false,
		Err:          nil,
		spinner:      s,
		format:       config.Format,
		responseCh:   make(chan tea.Msg),
	}
}

// Init starts the spinner's animation loop.
func (model StreamModel) Init() tea.Cmd {
	return tea.Batch(model.spinner.Tick, model.startStream())
}

// Update handles messages and returns the updated model and optional command.
func (model StreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width

		return model, nil

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			return model, tea.Quit
		}

	case StreamChunkMsg:
		model.content += string(msg)

		return model, listenForChunk(model.responseCh)

	case LLMDoneMsg:
		model.done = true

		return model, tea.Quit

	case StreamErrorMsg:
		model.Err = msg.Err
		model.done = true

		return model, tea.Quit

	default:
		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)

		return model, cmd
	}

	return model, nil
}

// View renders the current model state.
func (model StreamModel) View() tea.View {
	if model.done {
		return tea.NewView("") // Clear the view.
	}

	if model.content != "" {
		content, err := theme.RenderContent(model.content, model.format, true)
		if err != nil {
			model.logger.Error("content render failed", "error", err, "format", model.format)
			return tea.NewView("")
		}

		if model.format == "" {
			content = theme.WordWrap(model.width, content, theme.FgText)
		}

		return tea.NewView(content)
	}

	processingMessage := theme.FgAccent0.Render(theme.GlyphInfo+" processing") + model.spinner.View()

	return tea.NewView(processingMessage)
}

// Content returns the full model content with styling for normal text.
// JSON and Markdown output are returned raw.
func (model StreamModel) Content() string {
	if model.format == "json" || model.format == "markdown" {
		return model.content
	}

	return theme.WordWrap(model.width, model.content, theme.FgText)
}

func (model StreamModel) startStream() tea.Cmd {
	model.logger.Debug("transmitting to neural network", "model", model.model, "messages", len(model.messages))

	go func() {
		ch := model.responseCh
		defer close(ch)

		imageAnalysis, err := agent.AnalyseImages(model.ctx, model.url, model.visionModel, model.images, model.logger)
		if err != nil {
			ch <- StreamErrorMsg{Err: err}

			return
		}

		model.messages = append(model.messages, imageAnalysis...)

		model.messages, err = agent.RunToolLoop(model.ctx, model.toolRegistry, model.url, model.model, model.messages, model.logger)
		if err != nil {
			ch <- StreamErrorMsg{Err: err}

			return
		}

		_, err = llm.StreamChat(model.ctx, model.url, model.model, model.messages, nil, func(chunk string) {
			ch <- StreamChunkMsg(chunk)
		})

		if err != nil {
			ch <- StreamErrorMsg{Err: err}

			return
		}
	}()

	return listenForChunk(model.responseCh)
}
