package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInit(t *testing.T) {
	t.Run("initializes the model", func(t *testing.T) {
		t.Parallel()

		model := Model{}

		actualCmd := model.Init()

		if actualCmd != nil {
			t.Errorf("expected teaCmd to be nil, got %v", actualCmd)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("updates the model", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		msg := tea.KeyMsg{}

		returnedModel, actualCmd := model.Update(msg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		if actualModel.width != 0 {
			t.Errorf("expected model width to be 0, got %d", actualModel.width)
		}

		if actualModel.height != 0 {
			t.Errorf("expected model heigiht to be 0, got %d", actualModel.height)
		}

		if actualCmd != nil {
			t.Errorf("expected teaCmd to be nil, got %v", actualCmd)
		}
	})

	t.Run("handles window size message", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}

		returnedModel, actualCmd := model.Update(sizeMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		expectedWidth := 80
		expectedHeight := 24

		if actualModel.width != expectedWidth {
			t.Errorf("expected width %d, got %d", expectedWidth, actualModel.width)
		}

		if actualModel.height != expectedHeight {
			t.Errorf("expected height %d, got %d", expectedHeight, actualModel.height)
		}

		if actualCmd != nil {
			t.Errorf("expect teaCmd to be nil, got %v", actualCmd)
		}
	})
}

func TestView(t *testing.T) {
	t.Run("returns the view string", func(t *testing.T) {
		t.Parallel()

		model := Model{}

		actualView := model.View()

		if actualView != "" {
			t.Errorf("expected view to be empty, got %q", actualView)
		}
	})
}
