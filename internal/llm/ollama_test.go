package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/go-cmp/cmp"
)

var logger *log.Logger = log.New(io.Discard)

func TestNewOllamaClient(t *testing.T) {
	t.Run("creates new Ollama client with default", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response": "Hello, user!"}`))
		}))
		defer server.Close()

		client, err := NewOllamaClient("http://test.dev", "llama2", http.DefaultClient, logger)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if client.baseURL != "http://test.dev" {
			t.Errorf("expected baseURL to be 'http://test.dev', got '%s'", client.baseURL)
		}

		if client.defaultModel != "llama2" {
			t.Errorf("expected defaultModel to be 'llama2', got '%s'", client.defaultModel)
		}
	})

	t.Run("returns error for empty baseURL", func(t *testing.T) {
		t.Parallel()

		_, err := NewOllamaClient("", "llama2", http.DefaultClient, logger)

		if err == nil {
			t.Fatal("expected error for empty baseURL, got nil")
		}

		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("returns error for empty defaultModel", func(t *testing.T) {
		t.Parallel()

		_, err := NewOllamaClient("http://test.dev", "", http.DefaultClient, logger)

		if err == nil {
			t.Fatal("expected error for empty defaultModel, got nil")
		}

		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestStreamChat(t *testing.T) {
	t.Run("streams chat without error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Fatalf("expected response writer to support flushing")
			}

			chunks := []string{
				`{"message":{"content":"Hello, "},"done":false}`,
				`{"message":{"content":"user!"},"done":false}`,
				`{"message":{"content":""},"done":true}`,
			}

			for _, chunk := range chunks {
				_, _ = w.Write([]byte(chunk + "\n"))
				flusher.Flush()
			}
		}))

		defer server.Close()

		httpClient := &http.Client{
			Transport: server.Client().Transport,
		}

		client, err := NewOllamaClient(server.URL, "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		chatHistory := []ChatMessage{
			{Role: User, Content: "Hello, there!"},
		}

		var actual []string
		onToken := func(token string) {
			actual = append(actual, token)
		}

		err = client.StreamChat(context.Background(), chatHistory, onToken)
		if err != nil {
			t.Fatalf("expected no error calling StreamChat, got %v", err)
		}

		expected := []string{"Hello, ", "user!"}

		if !cmp.Equal(actual, expected) {
			t.Errorf("expected tokens %v, got %v", expected, actual)
		}
	})
}
