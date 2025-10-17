package llm

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewOllama(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		defaultModel string
		isError      bool
		err          error
	}{
		{
			name:         "creates a new Ollama client",
			baseURL:      "http://test.dev",
			defaultModel: "default:model",
		},
		{
			name:         "returns error for no base URL",
			defaultModel: "default:model",
			isError:      true,
			err:          ErrNoBaseURL,
		},
		{
			name:    "returns error for no default model",
			baseURL: "http://test.dev",
			isError: true,
			err:     ErrNoDefaultModel,
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

			if tt.isError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
			}

			if !tt.isError {
				if ollama.baseURL != tt.baseURL {
					t.Errorf("expected base URL %q, got %q", tt.baseURL, ollama.baseURL)
				}

				if ollama.defaultModel != tt.defaultModel {
					t.Errorf("expected default model %q, got %q", tt.defaultModel, ollama.defaultModel)
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	logger := log.New(io.Discard)

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		response := `{"response": "Hello, chummer!"}`

		_, _ = w.Write([]byte(response))
	}))

	defer httpServer.Close()

	httpClient := &http.Client{Transport: httpServer.Client().Transport}

	ollama, err := NewOllama(httpServer.URL, "test:model", httpClient, logger)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	systemPrompt := "test system prompt"
	userPrompt := "test user prompt"

	response := ollama.Generate(systemPrompt, userPrompt)

	expectedResponse := "Hello, chummer!"
	if response != expectedResponse {
		t.Errorf("expected response %q, got %q", expectedResponse, response)
	}
}
