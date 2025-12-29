package cmd

import (
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

func TestInitMessages(t *testing.T) {
	system := "system prompt"
	prompt := "user prompt"

	actual, err := initMessages(system, prompt, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := llm.ChatMessage{
		Role:    "system",
		Content: system,
	}

	if actual[0] != expected {
		t.Errorf("expected messages %v, got %v", expected, actual[0])
	}

	expected = llm.ChatMessage{
		Role:    "user",
		Content: prompt,
	}

	if actual[1] != expected {
		t.Errorf("expected messages %v, got %v", expected, actual[1])
	}
}
