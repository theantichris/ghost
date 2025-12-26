package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewOllama(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		isError bool
		err     error
	}{
		{
			name: "creates a new Ollama client",
			host: "http://test.dev",
		},
		{
			name:    "returns error for no host URL",
			isError: true,
			err:     ErrNoHostURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			ollama, err := NewOllama(tt.host, logger)

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
				if ollama.host != tt.host {
					t.Errorf("expected host URL %q, got %q", tt.host, ollama.host)
				}

				if ollama.generateURL != tt.host+"/api/generate" {
					t.Errorf("expected generate URL %q, got %q", tt.host+"/api/generate", ollama.generateURL)
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name         string
		systemPrompt string
		prompt       string
		images       []string
		httpStatus   int
		isError      bool
		err          error
	}{
		{
			name:         "returns response from API",
			systemPrompt: "test system prompt",
			prompt:       "test user prompt",
			images:       []string{},
			httpStatus:   http.StatusOK,
		},
		{
			name:         "returns API error",
			systemPrompt: "test system prompt",
			prompt:       "test user prompt",
			images:       []string{},
			httpStatus:   http.StatusBadRequest,
			isError:      true,
			err:          ErrOllama,
		},
		{
			name:         "returns 404 error",
			systemPrompt: "test system prompt",
			prompt:       "test user prompt",
			images:       []string{},
			httpStatus:   http.StatusNotFound,
			isError:      true,
			err:          ErrModelNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("cannot read request body: %v", err)
				}

				if len(tt.images) > 0 {
					if !strings.Contains(string(body), "vision:model") {
						t.Errorf("expected model to be vision:model, got %s", body)
					}
				} else {
					if !strings.Contains(string(body), "default:model") {
						t.Errorf("expected model to be default:model, got %s", body)
					}
				}

				response := `{"response": "Hello, chummer!"}`

				_, _ = w.Write([]byte(response))
			}))

			defer httpServer.Close()

			ollama, err := NewOllama(httpServer.URL, logger)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			var response string
			err = ollama.Generate(context.Background(), "default:model", tt.systemPrompt, tt.prompt, tt.images, func(chunk string) error {
				response += chunk
				return nil
			})

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

			ollama, err := NewOllama(httpServer.URL, logger)
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
		model      string
		httpStatus int
		isError    bool
	}{
		{
			name:       "returns ok for model found",
			model:      "default:model",
			httpStatus: http.StatusOK,
		},
		{
			name:       "returns not found error",
			model:      "default:model",
			httpStatus: http.StatusNotFound,
			isError:    true,
		},
		{
			name:       "returns API error",
			model:      "default:model",
			httpStatus: http.StatusBadGateway,
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

			ollama, err := NewOllama(httpServer.URL, logger)
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			err = ollama.Show(context.Background(), tt.model)

			if tt.isError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.isError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
