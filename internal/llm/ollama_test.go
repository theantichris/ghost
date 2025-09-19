package llm

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

type errorTransport struct{}

func (e *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

func TestNewOllamaClient(t *testing.T) {
	t.Run("creates new Ollama client with default", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response": "Hello, user!"}`))
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

		if !errors.Is(err, ErrURLEmpty) {
			t.Errorf("expected ErrURLEmpty, got %v", err)
		}
	})

	t.Run("returns error for empty defaultModel", func(t *testing.T) {
		t.Parallel()

		_, err := NewOllamaClient("http://test.dev", "", http.DefaultClient, logger)

		if err == nil {
			t.Fatal("expected error for empty defaultModel, got nil")
		}

		if !errors.Is(err, ErrModelEmpty) {
			t.Errorf("expected ErrModelEmpty, got %v", err)
		}
	})
}

func TestChat(t *testing.T) {
	t.Run("returns the response from Ollama", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": {"role": "assistant", "content": "Hello, user!"}}`))
		}))
		defer server.Close()

		httpClient := &http.Client{
			Transport: server.Client().Transport,
		}

		client, err := NewOllamaClient(server.URL, "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		response, err := client.Chat(context.Background(), "Hello, there!")
		if err != nil {
			t.Fatalf("expected no error calling Chat, got %v", err)
		}

		expectedResponse := "Hello, user!"
		if response != expectedResponse {
			t.Errorf("expected response '%s', got '%s'", expectedResponse, response)
		}
	})

	t.Run("returns error for missing message", func(t *testing.T) {
		t.Parallel()

		app, err := NewOllamaClient("http://test.dev", "llama2", http.DefaultClient, logger)
		if err != nil {
			t.Fatal("expected no error creating client, got", err)
		}

		_, err = app.Chat(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}

		if !errors.Is(err, ErrMessageEmpty) {
			t.Errorf("expected ErrMessageEmpty, got %v", err)
		}
	})

	t.Run("returns error for failed HTTP client creation", func(t *testing.T) {
		t.Parallel()

		client, err := NewOllamaClient(":", "llama2", http.DefaultClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		_, err = client.Chat(context.Background(), "Hello, there!")
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}

		if !strings.Contains(err.Error(), "missing protocol scheme") {
			t.Errorf("expected error containing 'missing protocol scheme', got %v", err)
		}
	})

	t.Run("returns error for context.DeadlineExceeded", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {} // Simulate a long processing time
		}))
		defer server.Close()

		httpClient := &http.Client{
			Transport: server.Client().Transport,
		}

		client, err := NewOllamaClient(server.URL, "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		_, err = client.Chat(ctx, "Hello, there!")
		if err == nil {
			t.Fatal("expected error for context deadline exceeded, got nil")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded error, got %v", err)
		}
	})

	t.Run("returns error for failed HTTP request", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{Transport: &errorTransport{}}

		client, err := NewOllamaClient("http://test.dev", "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		_, err = client.Chat(context.Background(), "Hello, there!")
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
			w.Write([]byte(`{"error": "Internal Server Error"}`))
		}))
		defer server.Close()

		httpClient := &http.Client{
			Transport: server.Client().Transport,
		}

		client, err := NewOllamaClient(server.URL, "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		_, err = client.Chat(context.Background(), "Hello, there!")
		if err == nil {
			t.Fatal("expected error for non-200 HTTP response, got nil")
		}

		if !strings.Contains(err.Error(), "non-2xx response: status=500 Internal Server Error") {
			t.Errorf("expected error containing 'non-2xx response: status=500 Internal Server Error', got %v", err)
		}
	})

	t.Run("returns error for invalid JSON response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		httpClient := &http.Client{
			Transport: server.Client().Transport,
		}

		client, err := NewOllamaClient(server.URL, "llama2", httpClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		_, err = client.Chat(context.Background(), "Hello, there!")
		if err == nil {
			t.Fatal("expected error for invalid JSON response, got nil")
		}

		if !strings.Contains(err.Error(), "failed to unmarshal response body: invalid character 'i' looking for beginning of value") {
			t.Errorf("expected error containing 'failed to unmarshal response body: invalid character 'i' looking for beginning of value', got %v", err)
		}
	})
}
