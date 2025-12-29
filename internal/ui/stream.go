package ui

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/theme"
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
	Content string        // Accumulated response content.
	done    bool          // Whether streaming has finished.
	Err     error         // Error if streaming failed.
	spinner spinner.Model // Animated spinner.
}

// NewStreamModel creates and returns StreamModel.
func NewStreamModel() StreamModel {
	s := spinner.New()
	s.Spinner = spinner.Ellipsis
	s.Style = theme.FgAccent0

	return StreamModel{
		width:   80,
		Content: "",
		done:    false,
		Err:     nil,
		spinner: s,
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
		// TODO: look into removing magic string
		if msg.String() == "ctrl+c" {
			return model, tea.Quit
		}

	case StreamChunkMsg:
		model.Content += string(msg)

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
	if model.Content != "" {
		wrappedContent := lipgloss.NewStyle().Width(model.width).Render(model.Content)
		return tea.NewView(wrappedContent)
	}

	processingMessage := theme.FgAccent0.Render("󱙝 processing") + model.spinner.View()

	return tea.NewView(processingMessage)
}
