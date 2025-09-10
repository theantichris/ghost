package llm

import "testing"

func TestNewOllamaClient(t *testing.T) {
	t.Run("creates a new Ollama client instance", func(t *testing.T) {
		client := NewOllamaClient("http://test.dev", "llama2")

		if client == nil {
			t.Fatal("expected client to be non-nil")
		}

		if client.baseURL != "http://test.dev" {
			t.Errorf("expected baseURL to be 'http://test.dev', got '%s'", client.baseURL)
		}

		if client.defaultModel != "llama2" {
			t.Errorf("expected defaultModel to be 'llama2', got '%s'", client.defaultModel)
		}
	})
}
