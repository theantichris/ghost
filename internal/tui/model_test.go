package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInit(t *testing.T) {
	t.Run("initializes the model", func(t *testing.T) {
		t.Parallel()

		model := Model{}

		actualTeaCmd := model.Init()

		if actualTeaCmd != nil {
			t.Errorf("expected teaCmd to be nil, got %v", actualTeaCmd)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("updates the model", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		msg := tea.KeyMsg{}

		actualTeaModel, actualTeaCmd := model.Update(msg)

		if actualTeaModel != nil {
			t.Errorf("expected model to be nil, got %v", actualTeaModel)
		}

		if actualTeaCmd != nil {
			t.Errorf("expected teaCmd to be nil, got %v", actualTeaCmd)
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
