package tui

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theantichris/ghost/internal/llm"
)

func TestNewModel(t *testing.T) {
	t.Run("creates a new model with dependencies and system prompt", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		systemPrompt := "This is the system prompt."

		actualModel := NewModel(systemPrompt, logger)

		if actualModel.logger == nil {
			t.Error("expected logger to be set")
		}

		expectedChatLength := 2

		if len(actualModel.chatHistory) != expectedChatLength {
			t.Fatalf("expected chat to contain %d items, got %d", expectedChatLength, len(actualModel.chatHistory))
		}

		actualSystemPrompt := actualModel.chatHistory[0]

		if actualSystemPrompt.Role != llm.SystemRole {
			t.Errorf("expected system prompt to have user role, got %q", actualSystemPrompt.Role)
		}

		if actualSystemPrompt.Content != systemPrompt {
			t.Errorf("expected system prompt, got %q", actualSystemPrompt.Content)
		}

		actualGreetingPrompt := actualModel.chatHistory[1]

		if actualGreetingPrompt.Role != llm.SystemRole {
			t.Errorf("expected greeting prompt to have user role, got %q", actualGreetingPrompt.Role)
		}

		if actualGreetingPrompt.Content != "Greet the user." {
			t.Errorf("expected greeting prompt, got %q", actualGreetingPrompt.Content)
		}

		if actualModel.input != "" {
			t.Errorf("expected empty input, got %q", actualModel.input)
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("returns command to send greeting", func(t *testing.T) {
		t.Parallel()

		model := Model{
			chatHistory: []llm.ChatMessage{
				{Role: llm.SystemRole, Content: "test system prompt"},
				{Role: llm.SystemRole, Content: "test greeting prompt"},
			},
		}

		actualCmd := model.Init()

		if actualCmd == nil {
			t.Fatal("expected command to send greeting, got nil")
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

	t.Run("handles stream chunk message", func(t *testing.T) {
		t.Parallel()

		model := Model{}
		msg := streamingChunkMsg{content: "Hello"}

		returnedModel, _ := model.Update(msg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		expectedCurrentMsg := "Hello"

		if actualModel.currentMsg != "Hello" {
			t.Errorf("expected current message to be %q, got %q", expectedCurrentMsg, actualModel.currentMsg)
		}

		if !actualModel.streaming {
			t.Error("expected streaming to be true, got false")
		}
	})

	t.Run("appends multiple stream chunks", func(t *testing.T) {
		t.Parallel()

		model := Model{currentMsg: "Hello", streaming: true}
		msg := streamingChunkMsg{content: " world"}

		returnedModel, _ := model.Update(msg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		expectedCurrentMsg := "Hello world"

		if actualModel.currentMsg != expectedCurrentMsg {
			t.Errorf("expected current message to be %q, got %q", expectedCurrentMsg, actualModel.currentMsg)
		}
	})

	t.Run("handles stream complete message", func(t *testing.T) {
		t.Parallel()

		model := Model{
			currentMsg: "Hello, how can I help?",
			streaming:  true,
		}
		msg := streamCompleteMsg{}

		returnedModel, _ := model.Update(msg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		if actualModel.streaming {
			t.Error("expected streaming to be false, got true")
		}

		if actualModel.currentMsg != "" {
			t.Errorf("expected currentMsg to be empty, got %q", actualModel.currentMsg)
		}

		expectedMessagesLength := 1

		if len(actualModel.messages) != expectedMessagesLength {
			t.Errorf("expected messages length %d, got %d", expectedMessagesLength, len(actualModel.messages))
		}

		expectedMessage := "Hello, how can I help?"
		actualMessage := actualModel.messages[0]

		if actualMessage != expectedMessage {
			t.Errorf("expected message %q, got %q", expectedMessage, actualMessage)
		}

		expectedHistoryLength := 1
		actualHistoryLength := len(actualModel.chatHistory)

		if actualHistoryLength != expectedHistoryLength {
			t.Errorf("expected history length %d, got %d", expectedHistoryLength, actualHistoryLength)
		}

		actualChatMessage := actualModel.chatHistory[0]

		if actualChatMessage.Role != llm.AssistantRole {
			t.Errorf("expected role to be %q, got %q", llm.AssistantRole, actualChatMessage.Role)
		}

		if actualChatMessage.Content != expectedMessage {
			t.Errorf("expected chat message %q, got %q", expectedMessage, actualChatMessage.Content)
		}
	})

	t.Run("handles stream error message", func(t *testing.T) {
		t.Parallel()

		testError := errors.New("connection failed")

		model := Model{
			currentMsg: "Hello",
			streaming:  true,
		}
		msg := streamErrorMsg{err: testError}

		returnedModel, _ := model.Update(msg)

		actualModel, ok := returnedModel.(Model)
		if !ok {
			t.Fatal("expected model to be of type Model")
		}

		if actualModel.streaming {
			t.Error("expected streaming to be false, got true")
		}

		if actualModel.err == nil {
			t.Fatal("expected error to be set, got nil")
		}

		if !errors.Is(actualModel.err, testError) {
			t.Errorf("expected error %v, got %v", testError, actualModel.err)
		}

		if actualModel.currentMsg != "" {
			t.Errorf("expected currentMsg to be cleared, got %q", actualModel.currentMsg)
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

func TestSendChatRequest(t *testing.T) {
	t.Run("sends chat request and returns stream messages", func(t *testing.T) {
		t.Parallel()

		tokens := []string{"Hello", " world"}
		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, messages []llm.ChatMessage, onToken func(string)) error {
				for _, token := range tokens {
					onToken(token)
				}

				return nil
			},
		}

		model := Model{
			llmClient: mockClient,
			chatHistory: []llm.ChatMessage{
				{Role: llm.SystemRole, Content: "test"},
			},
		}

		msg := model.sendChatRequest()

		if msg == nil {
			t.Fatal("expected message to be returned, not nil")
		}
	})
}
