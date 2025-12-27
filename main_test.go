package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlmjohnson/requests"
)

func TestGetPrompt(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
		err      error
	}{
		{
			name:     "returns prompt",
			args:     []string{"ghost", "tell me a joke"},
			expected: "tell me a joke",
		},
		{
			name:    "returns error for no prompt",
			args:    []string{"ghost"},
			wantErr: true,
			err:     errPromptNotDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getPrompt(tt.args)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if actual != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, actual)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
			}
		})
	}
}

func TestCreateMessages(t *testing.T) {
	system := "system prompt"
	prompt := "user prompt"

	messages := createMessages(system, prompt)

	expected := chatMessage{
		Role:    "system",
		Content: "system prompt",
	}

	if messages[0] != expected {
		t.Errorf("expected messages %v, got %v", expected, messages[0])
	}

	expected = chatMessage{
		Role:    "user",
		Content: "user prompt",
	}

	if messages[1] != expected {
		t.Errorf("expected messages %v, got %v", expected, messages[0])
	}
}

func TestGetChatResponse(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		messages     []chatMessage
		chatResponse chatResponse
		serverStatus int
		wantContent  string
		wantErr      bool
	}{
		{
			name:  "successful response",
			model: "test-model",
			messages: []chatMessage{
				{Role: "system", Content: "You are a test assistant."},
				{Role: "user", Content: "Hello"},
			},
			chatResponse: chatResponse{
				Message: chatMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
			},
			serverStatus: http.StatusOK,
			wantContent:  "Hello! How can I help you?",
			wantErr:      false,
		},
		{
			name:  "empty response content",
			model: "test-model",
			messages: []chatMessage{
				{Role: "user", Content: "test"},
			},
			chatResponse: chatResponse{
				Message: chatMessage{
					Role:    "assistant",
					Content: "",
				},
			},
			serverStatus: http.StatusOK,
			wantContent:  "",
			wantErr:      false,
		},
		{
			name:  "server error",
			model: "test-model",
			messages: []chatMessage{
				{Role: "user", Content: "test"},
			},
			chatResponse: chatResponse{},
			serverStatus: http.StatusInternalServerError,
			wantContent:  "",
			wantErr:      true,
		},
		{
			name:  "multiple messages in history",
			model: "test-model",
			messages: []chatMessage{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "First question"},
				{Role: "assistant", Content: "First answer"},
				{Role: "user", Content: "Second question"},
			},
			chatResponse: chatResponse{
				Message: chatMessage{
					Role:    "assistant",
					Content: "Second answer here",
				},
			},
			serverStatus: http.StatusOK,
			wantContent:  "Second answer here",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}

				if r.URL.Path != "/chat" {
					t.Errorf("expected /chat path, got %s", r.URL.Path)
				}

				var receivedRequest chatRequest
				if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}

				if receivedRequest.Model != tt.model {
					t.Errorf("expected model %q, got %q", tt.model, receivedRequest.Model)
				}

				if len(receivedRequest.Messages) != len(tt.messages) {
					t.Errorf("expected %d messages, got %d", len(tt.messages), len(receivedRequest.Messages))
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					if err := json.NewEncoder(w).Encode(tt.chatResponse); err != nil {
						t.Fatalf("failed to encode response: %v", err)
					}
				}
			}))
			defer server.Close()

			ctx := context.Background()

			request := chatRequest{
				Model:    tt.model,
				Stream:   false,
				Messages: tt.messages,
			}

			var chatResponse chatResponse

			err := requests.
				URL(server.URL + "/chat").
				BodyJSON(&request).
				ToJSON(&chatResponse).
				Fetch(ctx)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if chatResponse.Message.Content != tt.wantContent {
					t.Errorf("expected %q, got %q", tt.wantContent, chatResponse.Message.Content)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
		})
	}
}
