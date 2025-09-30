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
	"github.com/google/go-cmp/cmp"
)

var logger *log.Logger = log.New(io.Discard)

type errorTransport struct{}

func (e *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

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

func TestChat(t *testing.T) {
	t.Run("returns the response from Ollama", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": {"role": "assistant", "content": "Hello, user!"}}`))
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

		actual, err := client.Chat(context.Background(), chatHistory)
		if err != nil {
			t.Fatalf("expected no error calling Chat, got %v", err)
		}

		expected := ChatMessage{Role: Assistant, Content: "Hello, user!"}

		if !cmp.Equal(actual, expected) {
			t.Errorf("expected response to be %v, got %v", expected, actual)
		}
	})

	t.Run("returns error for empty chat history", func(t *testing.T) {
		t.Parallel()

		app, err := NewOllamaClient("http://test.dev", "llama2", http.DefaultClient, logger)
		if err != nil {
			t.Fatal("expected no error creating client, got", err)
		}

		_, err = app.Chat(context.Background(), []ChatMessage{})
		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}

		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("returns error for failed HTTP client creation", func(t *testing.T) {
		t.Parallel()

		client, err := NewOllamaClient(":", "llama2", http.DefaultClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		chatHistory := []ChatMessage{
			{Role: User, Content: "Hello, there!"},
		}

		_, err = client.Chat(context.Background(), chatHistory)
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}

		if !strings.Contains(err.Error(), "missing protocol scheme") {
			t.Errorf("expected error containing 'missing protocol scheme', got %v", err)
		}
	})

	t.Run("returns error for failed HTTP request", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{Transport: &errorTransport{}}

		client, err := NewOllamaClient("http://test.dev", "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		chatHistory := []ChatMessage{
			{Role: User, Content: "Hello, there!"},
		}

		_, err = client.Chat(context.Background(), chatHistory)
		if err == nil {
			t.Fatal("expected error for failed HTTP request, got nil")
		}

		if !strings.Contains(err.Error(), "forced transport error") {
			t.Errorf("expected error containing 'forced transport error', got %v", err)
		}
	})

	t.Run("returns error for non-200 HTTP response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal Server Error"}`))
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

		_, err = client.Chat(context.Background(), chatHistory)
		if err == nil {
			t.Fatal("expected error for non-200 HTTP response, got nil")
		}

		if !strings.Contains(err.Error(), "failed to process response: status=500 Internal Server Error") {
			t.Errorf("expected error containing 'failed to process response: status=500 Internal Server Error', got %v", err)
		}
	})

	t.Run("returns error for invalid JSON response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid json`))
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

		_, err = client.Chat(context.Background(), chatHistory)
		if err == nil {
			t.Fatal("expected error for invalid JSON response, got nil")
		}

		if !strings.Contains(err.Error(), "failed to process response: invalid character 'i' looking for beginning of value") {
			t.Errorf("expected error containing 'failed to process response: invalid character 'i' looking for beginning of value', got %v", err)
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
