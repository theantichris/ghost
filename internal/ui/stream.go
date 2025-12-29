package ui

import tea "charm.land/bubbletea/v2"

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
	Content string // Accumulated response content.
	done    bool   // Whether streaming has finished.
	Err     error  // Error if streaming failed.
}

// NewStreamModel creates and returns StreamModel.
func NewStreamModel() StreamModel {
	return StreamModel{
		Content: "",
		done:    false,
		Err:     nil,
	}
}

// Init returns the initial command to execute on startup.
func (model StreamModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns the updated model and optional command.
func (model StreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
	}

	return model, nil
}

// View renders the current model state.
func (model StreamModel) View() tea.View {
	if model.Content != "" {
		return tea.NewView(model.Content)
	}

	if model.Err != nil {
		return tea.NewView("󱙝 error: " + model.Err.Error())
	}

	return tea.NewView("󱙝 processing...")
}
