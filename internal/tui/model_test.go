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
	t.Run("handles window size message", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}

		returnedModel, _ := model.Update(sizeMsg)

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
	})

	t.Run("handles regular key press", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		expectedInput := "h"

		if actualModel.input != expectedInput {
			t.Errorf("expected model input to be %q, got %q", expectedInput, actualModel.input)
		}
	})

	t.Run("handles backspace key", func(t *testing.T) {
		t.Parallel()

		model := Model{input: "hello"}
		keyMsg := tea.KeyMsg{Type: tea.KeyBackspace}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		expectedInput := "hell"

		if actualModel.input != expectedInput {
			t.Errorf("expected model input to be %q, got %q", expectedInput, actualModel.input)
		}
	})

	t.Run("handles ctrl+d to exit", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		keyMsg := tea.KeyMsg{Type: tea.KeyCtrlD}

		returnedModel, actualCmd := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		if !actualModel.exiting {
			t.Errorf("expected model exiting to be true, got false")
		}

		quitMsg := actualCmd()
		if _, ok := quitMsg.(tea.QuitMsg); !ok {
			t.Errorf("expected command to return tea.QuitMsg, got %T", quitMsg)
		}
	})

	t.Run("handles ctrl+c to exit", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}

		returnedModel, actualCmd := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		if !actualModel.exiting {
			t.Errorf("expected model exiting to be true, got false")
		}

		quitMsg := actualCmd()
		if _, ok := quitMsg.(tea.QuitMsg); !ok {
			t.Errorf("expected command to return tea.QuitMsg, got %T", quitMsg)
		}
	})

	t.Run("enter key clears input", func(t *testing.T) {
		t.Parallel()

		model := Model{input: "hello"}
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		expectedInput := ""

		if actualModel.input != expectedInput {
			t.Errorf("expected input to be cleared, got %q", actualModel.input)
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
