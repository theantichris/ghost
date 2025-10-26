package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewOllama(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		isError bool
		err     error
	}{
		{
			name: "creates a new Ollama client",
			config: Config{
				Host:         "http://test.dev",
				DefaultModel: "default:model",
				VisionModel:  "vision:model",
			},
		},
		{
			name: "returns error for no host URL",
			config: Config{
				DefaultModel: "default:model",
				VisionModel:  "vision:model",
			},
			isError: true,
			err:     ErrNoHostURL,
		},
		{
			name: "returns error for no default model",
			config: Config{
				Host:        "http://test.dev",
				VisionModel: "vision:model",
			},
			isError: true,
			err:     ErrNoDefaultModel,
		},
		{
			name: "returns error for no vision model",
			config: Config{
				Host:         "http://test.dev",
				DefaultModel: "default:model",
			},
			isError: true,
			err:     ErrNoVisionModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			ollama, err := NewOllama(tt.config, logger)

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
				if ollama.host != tt.config.Host {
					t.Errorf("expected host URL %q, got %q", tt.config.Host, ollama.host)
				}

				if ollama.generateURL != tt.config.Host+"/api/generate" {
					t.Errorf("expected generate URL %q, got %q", tt.config.Host+"/api/generate", ollama.generateURL)
				}

				if ollama.defaultModel != tt.config.DefaultModel {
					t.Errorf("expected default model %q, got %q", tt.config.DefaultModel, ollama.defaultModel)
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		isError    bool
		err        error
	}{
		{
			name:       "returns response from API",
			httpStatus: http.StatusOK,
		},
		{
			name:       "returns API error",
			httpStatus: http.StatusNotFound,
			isError:    true,
			err:        ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)

				response := `{"response": "Hello, chummer!"}`

				_, _ = w.Write([]byte(response))
			}))

			defer httpServer.Close()

			config := Config{
				Host:         httpServer.URL,
				DefaultModel: "test:model",
				VisionModel:  "vision:model",
			}

			ollama, err := NewOllama(config, logger)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			systemPrompt := "test system prompt"
			userPrompt := "test user prompt"

			response, err := ollama.Generate(context.Background(), systemPrompt, userPrompt, []string{})

			if !tt.isError && err != nil {
				t.Fatalf("expected no error, got %v", err)
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
				expectedResponse := "Hello, chummer!"
				if response != expectedResponse {
					t.Errorf("expected response %q, got %q", expectedResponse, response)
				}
			}
		})
	}
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		isError    bool
		err        error
	}{
		{
			name:       "returns API version",
			httpStatus: http.StatusOK,
		},
		{
			name:       "returns API error",
			httpStatus: http.StatusNotFound,
			isError:    true,
			err:        ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)

				response := `{"version": "0.12.6"}`

				_, _ = w.Write([]byte(response))
			}))

			defer httpServer.Close()

			config := Config{
				Host:         httpServer.URL,
				DefaultModel: "test:model",
				VisionModel:  "vision:model",
			}

			ollama, err := NewOllama(config, logger)
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			version, err := ollama.Version(context.Background())
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			if !tt.isError {
				if version != "0.12.6" {
					t.Errorf("expected version 0.12.6, got %q", version)
				}
			}
		})
	}
}

func TestShow(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		isError    bool
	}{
		{
			name:       "returns ok for model found",
			httpStatus: http.StatusOK,
		},
		{
			name:       "returns API error",
			httpStatus: http.StatusNotFound,
			isError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)

				response := `{"version": "0.12.6"}`

				_, _ = w.Write([]byte(response))
			}))

			defer httpServer.Close()

			config := Config{
				Host:         httpServer.URL,
				DefaultModel: "test:model",
				VisionModel:  "vision:model",
			}

			ollama, err := NewOllama(config, logger)
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			err = ollama.Show(context.Background())

			if tt.isError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.isError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
