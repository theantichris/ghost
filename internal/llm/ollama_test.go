package llm

import (
	"io"
	"net/http"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewOllama(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		defaultModel string
		isError      bool
	}{
		{
			name:         "creates a new Ollama client",
			baseURL:      "http://test.dev",
			defaultModel: "default:model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := http.DefaultClient
			logger := log.New(io.Discard)

			ollama, err := NewOllama(tt.baseURL, tt.defaultModel, httpClient, logger)

			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			if ollama.baseURL != tt.baseURL {
				t.Errorf("expected base URL %q, got %q", tt.baseURL, ollama.baseURL)
			}

			if ollama.defaultModel != tt.defaultModel {
				t.Errorf("expected default model %q, got %q", tt.defaultModel, ollama.defaultModel)
			}
		})
	}
}
