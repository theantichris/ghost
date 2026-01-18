package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamChat(t *testing.T) {
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
			model: "test:model",
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
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
			err:            ErrModelNotFound,
		},
		{
			name:  "returns error for unexpected status",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusInternalServerError,
			wantErr:        true,
			err:            ErrUnexpectedStatus,
		},
		{
			name:  "returns error for malformed JSON",
			model: "test:model",
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

			got, err := StreamChat(context.Background(), server.URL, tt.model, tt.messages, nil, onChunk)

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

func TestChat(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		messages       []ChatMessage
		tools          []Tool
		mockStatusCode int
		mockResponse   string
		wantContent    string
		wantToolCalls  int
		wantErr        bool
		err            error
	}{
		{
			name:  "successful chat",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"message":{"role":"assistant","content":"Hello there!"}}`,
			wantContent:    "Hello there!",
		},
		{
			name:  "chat with tool call response",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Search for Go news"},
			},
			tools: []Tool{
				{
					Type: "function",
					Function: ToolFunction{
						Name:        "web_search",
						Description: "Search the web",
						Parameters:  ToolParameters{Type: "object"},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"web_search","arguments":{"query":"Go news"}}}]}}`,
			wantContent:    "",
			wantToolCalls:  1,
		},
		{
			name:  "returns error for model not found",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusNotFound,
			mockResponse:   `{"error":"model not found"}`,
			wantErr:        true,
			err:            ErrModelNotFound,
		},
		{
			name:  "returns error for unexpected status",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"error":"internal server error"}`,
			wantErr:        true,
			err:            ErrUnexpectedStatus,
		},
		{
			name:  "returns error for model not supporting tools",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hello"},
			},
			tools: []Tool{
				{Type: "function", Function: ToolFunction{Name: "test"}},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"error":"dolphin-mistral does not support tools"}`,
			wantErr:        true,
			err:            ErrToolSupport,
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

			got, err := Chat(context.Background(), server.URL, tt.model, tt.messages, tt.tools)

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

			if got.Content != tt.wantContent {
				t.Errorf("Chat() content = %v, want %v", got.Content, tt.wantContent)
			}

			if len(got.ToolCalls) != tt.wantToolCalls {
				t.Errorf("Chat() tool_calls count = %v, want %v", len(got.ToolCalls), tt.wantToolCalls)
			}
		})
	}
}

func TestAnalyzeImages(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		messages       []ChatMessage
		mockStatusCode int
		mockResponse   string
		wantContent    string
		wantErr        bool
		err            error
	}{
		{
			name:  "successful image analysis",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Describe this image", Images: []string{"base64encodedimage"}},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"message":{"role":"assistant","content":"This image shows a cat."}}`,
			wantContent:    "This image shows a cat.",
		},
		{
			name:  "returns error for model not found",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Describe this image"},
			},
			mockStatusCode: http.StatusNotFound,
			mockResponse:   `{"error":"model not found"}`,
			wantErr:        true,
			err:            ErrModelNotFound,
		},
		{
			name:  "returns error for unexpected status",
			model: "test:model",
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Describe this image"},
			},
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"error":"internal server error"}`,
			wantErr:        true,
			err:            ErrUnexpectedStatus,
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

			got, err := AnalyzeImages(context.Background(), server.URL, tt.model, tt.messages)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("AnalyzeImages() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("AnalyzeImages() err = %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("AnalyzeImages() error = %v, want no error", err)
			}

			if got.Role != RoleAssistant {
				t.Errorf("AnalyzeImages() role = %v, want %v", got.Role, RoleAssistant)
			}

			if got.Content != tt.wantContent {
				t.Errorf("AnalyzeImages() content = %v, want %v", got.Content, tt.wantContent)
			}
		})
	}
}
