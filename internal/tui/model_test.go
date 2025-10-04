package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theantichris/ghost/internal/llm"
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

	t.Run("enter key adds message to chat history", func(t *testing.T) {
		t.Parallel()

		model := Model{input: "hello"}
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type model")
		}

		expectedHistoryLength := 1

		if len(actualModel.chatHistory) != expectedHistoryLength {
			t.Errorf("expected chat history length %d, got %d", expectedHistoryLength, len(actualModel.chatHistory))
		}

		expectedRole := llm.UserRole
		expectedContent := "hello"

		if actualModel.chatHistory[0].Role != expectedRole {
			t.Errorf("expected role %q, got %q", expectedRole, actualModel.chatHistory[0].Role)
		}

		if actualModel.chatHistory[0].Content != expectedContent {
			t.Errorf("expected content %q, got %q", expectedContent, actualModel.chatHistory[0].Content)
		}
	})

	t.Run("/bye command quits chat", func(t *testing.T) {
		t.Parallel()

		model := Model{input: "/bye"}
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		if !actualModel.exiting {
			t.Error("expected model exiting to be true, got false")
		}

		expectedContent := "Goodbye!"

		if len(actualModel.chatHistory) != 1 {
			t.Errorf("expected chat history length %d, got %d", 1, len(actualModel.chatHistory))
		}

		if actualModel.chatHistory[0].Content != expectedContent {
			t.Errorf("expected message content %q, got %q", expectedContent, actualModel.chatHistory[0].Content)
		}
	})

	t.Run("/exit command quits chat", func(t *testing.T) {
		t.Parallel()

		model := Model{input: "/exit"}
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

		returnedModel, _ := model.Update(keyMsg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		if !actualModel.exiting {
			t.Error("expected model exiting to be true, got false")
		}

		expectedContent := "Goodbye!"

		if len(actualModel.chatHistory) != 1 {
			t.Errorf("expected chat history length %d, got %d", 1, len(actualModel.chatHistory))
		}

		if actualModel.chatHistory[0].Content != expectedContent {
			t.Errorf("expected message content %q, got %q", expectedContent, actualModel.chatHistory[0].Content)
		}
	})
}

func TestView(t *testing.T) {
	t.Run("renders input field", func(t *testing.T) {
		t.Parallel()

		model := Model{
			input:  "hello",
			width:  80,
			height: 24,
		}

		actualView := model.View()

		expectedInput := "hello"

		if !strings.Contains(actualView, expectedInput) {
			t.Errorf("expected view to contain input %q, got: %q", expectedInput, actualView)
		}
	})

	t.Run("renders separator line", func(t *testing.T) {
		t.Parallel()

		model := Model{
			width:  80,
			height: 24,
		}

		actualView := model.View()

		expectedSeparator := strings.Repeat("─", 80)

		if !strings.Contains(actualView, expectedSeparator) {
			t.Errorf("expected view to contain separator line of 80 characters, got %q", actualView)
		}
	})

	t.Run("renders chat messages", func(t *testing.T) {
		t.Parallel()

		model := Model{
			messages: []string{"Hello, how are you?", "I'm doing great!"},
			width:    80,
			height:   24,
		}

		actualView := model.View()

		expectedMessage1 := "Hello, how are you?"
		expectedMessage2 := "I'm doing great!"

		if !strings.Contains(actualView, expectedMessage1) {
			t.Errorf("expected view to contain %q, got %q", expectedMessage1, actualView)
		}

		if !strings.Contains(actualView, expectedMessage2) {
			t.Errorf("expected view to contain %q, got %q", expectedMessage2, actualView)
		}
	})
}
