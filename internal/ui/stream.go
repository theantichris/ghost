package ui

import tea "charm.land/bubbletea/v2"

// StreamModel handles the UI for streaming LLM responses.
type StreamModel struct {
	content string // Accumulated response content.
	done    bool   // Whether streaming has finished.
}

// NewStreamModel creates and returns StreamModel.
func NewStreamModel() StreamModel {
	return StreamModel{
		content: "",
		done:    false,
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
	}

	return model, nil
}

// View renders the current model state.
func (model StreamModel) View() tea.View {
	if model.done {
		return tea.NewView(model.content)
	}

	return tea.NewView("󱙝 processing...")
}
