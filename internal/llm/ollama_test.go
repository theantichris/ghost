package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChat(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		messages       []ChatMessage
		mockStatusCode int
		mockResponse   string
		wantContent    string
		wantChunkCount int
		wantErr        bool
		err            error
	}{
		{
			name:  "successful streaming chat",
			model: "llama2",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusOK,
			mockResponse: `{"message":{"role":"assistant","content":"Hello"}}
{"message":{"role":"assistant","content":" there"}}
{"message":{"role":"assistant","content":"!"}}
`,
			wantContent:    "Hello there!",
			wantChunkCount: 3,
		},
		{
			name:  "returns error for model not found",
			model: "nonexistent",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
			err:            ErrModelNotFound,
		},
		{
			name:  "returns error for unexpected status",
			model: "llama2",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusInternalServerError,
			wantErr:        true,
			err:            ErrUnexpectedStatus,
		},
		{
			name:  "returns error for malformed JSON",
			model: "llama2",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"invalid json`,
			wantErr:        true,
			err:            ErrDecodeChunk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/chat" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			var chunks []string
			onChunk := func(content string) {
				chunks = append(chunks, content)
			}

			got, err := Chat(context.Background(), server.URL, tt.model, tt.messages, onChunk)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Chat() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("Chat() err = %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Chat() error = %v, want no error", err)
			}

			if got.Role != RoleAssistant {
				t.Errorf("Chat() role = %v, want %v", got.Role, RoleAssistant)
			}

			if got.Content != tt.wantContent {
				t.Errorf("Chat() content = %v, want %v", got.Content, tt.wantContent)
			}

			if len(chunks) != tt.wantChunkCount {
				t.Errorf("Chat() chunk count = %v, want %v", len(chunks), tt.wantChunkCount)
			}

			concatenated := strings.Join(chunks, "")
			if concatenated != tt.wantContent {
				t.Errorf("Chat() concatenated chunks = %v, want %v", concatenated, tt.wantContent)
			}
		})
	}
}
