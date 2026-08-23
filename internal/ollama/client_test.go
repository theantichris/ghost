package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "accepts absolute URL",
			baseURL: DefaultURL,
		},
		{
			name:    "rejects missing scheme",
			baseURL: "localhost:11434/api",
			wantErr: true,
		},
		{
			name:    "rejects empty URL",
			baseURL: "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewClient(test.baseURL, nil)

			if test.wantErr {
				if err == nil {
					t.Fatal("NewClient() error = nil, want an error")
				}

				return
			}

			if err != nil {
				t.Fatalf("NewClient() err = %v, want nil", err)
			}

			if client == nil {
				t.Fatal("NewClient() client = nil, want a client")
			}
		})
	}
}

func TestClientChat(t *testing.T) {
	messages := []Message{
		{
			Role:    "user",
			Content: "Identify yourself.",
		},
	}

	wantMessage := Message{
		Role:    "assistant",
		Content: "Ghost online.",
	}

	tests := []struct {
		name            string
		handler         http.HandlerFunc
		wantMessage     Message
		wantErrContains string
	}{
		{
			name: "sends chat request and returns assistant message",
			handler: func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Method != http.MethodPost {
					t.Errorf("request method = %q, want %q", request.Method, http.MethodPost)
				}

				if request.URL.Path != "/api/chat" {
					t.Errorf("request path = %q, want %q", request.URL.Path, "/api/chat")
				}

				if got := request.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want %q", got, "application/json")
				}

				var gotRequest chatRequest

				if err := json.NewDecoder(request.Body).Decode(&gotRequest); err != nil {
					t.Errorf("decode request: %v", err)
					return
				}

				if gotRequest.Model != "qwen3:8b" {
					t.Errorf("request model = %q, want %q", gotRequest.Model, "qwen3:8b")
				}

				if gotRequest.Stream {
					t.Error("request stream = true, want false")
				}

				if len(gotRequest.Messages) != 1 || gotRequest.Messages[0] != messages[0] {
					t.Errorf("request messages = %#v, want %#v", gotRequest.Messages, messages)
				}

				writer.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(writer).Encode(chatResponse{Message: wantMessage}); err != nil {
					t.Errorf("encode response: %v", err)
				}
			},
			wantMessage: wantMessage,
		},
		{
			name: "returns Ollama API error",
			handler: func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusNotFound)

				if err := json.NewEncoder(writer).Encode(errorResponse{Error: "model not found"}); err != nil {
					t.Errorf("encode error response: %v", err)
				}
			},
			wantErrContains: "404 Not Found: model not found",
		},
		{
			name: "returns malformed response error",
			handler: func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte("{"))
			},
			wantErrContains: "decode Ollama chat response",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			t.Cleanup(server.Close)

			client, err := NewClient(server.URL+"/api", server.Client())
			if err != nil {
				t.Fatalf("NewClient() error = %v, want nil", err)
			}

			gotMessage, err := client.Chat(context.Background(), "qwen3:8b", messages)

			if test.wantErrContains != "" {
				if err == nil {
					t.Fatal("Chat() error = nil, want an error")
				}

				if !strings.Contains(err.Error(), test.wantErrContains) {
					t.Errorf("Chat() error = %q, want it to contain %q", err, test.wantErrContains)
				}

				return
			}

			if err != nil {
				t.Fatalf("Chat() error = %v, want nil", err)
			}

			if gotMessage != test.wantMessage {
				t.Errorf("Chat() message = %#v, want %#v", gotMessage, test.wantMessage)
			}
		})
	}
}
