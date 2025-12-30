package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/theantichris/ghost/theme"
)

var quitKeys = key.NewBinding(
	key.WithKeys("ctrl+c"),
)

// StreamChunkMsg represents a chunk of text received from the LLM.
type StreamChunkMsg string

// StreamDoneMsg signals that streaming has completed.
type StreamDoneMsg struct{}

// StreamErrorMsg signals an error occurred during streaming.
type StreamErrorMsg struct {
	Err error
}

// StreamModel handles the UI for streaming LLM responses.
type StreamModel struct {
	width   int           // Terminal width
	content string        // Accumulated response content.
	done    bool          // Whether streaming has finished.
	Err     error         // Error if streaming failed.
	spinner spinner.Model // Animated spinner.
	format  string        // Format for output.
}

// NewStreamModel creates and returns StreamModel.
func NewStreamModel(format string) StreamModel {
	s := spinner.New()
	s.Spinner = spinner.Ellipsis
	s.Style = theme.FgAccent0

	return StreamModel{
		width:   80,
		content: "",
		done:    false,
		Err:     nil,
		spinner: s,
		format:  format,
	}
}

// Init starts the spinner's animation loop.
func (model StreamModel) Init() tea.Cmd {
	return model.spinner.Tick
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

		return model, nil

	case StreamDoneMsg:
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
	// Only stream normal text output.
	if model.content != "" && model.format != "json" && model.format != "markdown" {
		wrappedContent := theme.WordWrap(model.width, model.content)

		return tea.NewView(wrappedContent)
	}

	processingMessage := theme.FgAccent0.Render("󱙝 processing") + model.spinner.View()

	return tea.NewView(processingMessage)
}

// Content returns the full model content with styling
func (model StreamModel) Content() string {
	if model.format == "json" {
		return model.content
	}

	if model.format == "markdown" {
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithStyles(theme.CyberpunkTheme()),
			glamour.WithWordWrap(model.width),
		)

		out, _ := renderer.Render(model.content)

		return out
	}

	return theme.WordWrap(model.width, model.content)
}
