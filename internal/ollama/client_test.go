package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/theantichris/ghost/v4/internal/conversation"
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

func TestClientChatStream(t *testing.T) {
	callbackErr := errors.New("output unavailable")
	messages := []conversation.Message{
		{Role: conversation.RoleSystem, Content: "You are Ghost."},
		{Role: conversation.RoleUser, Content: "Identify yourself."},
	}

	tests := []struct {
		name            string
		statusCode      int
		body            string
		callbackErr     error
		wantChunks      []string
		wantErr         error
		wantErrContains string
	}{
		{
			name: "emits ordered content including whitespace and final frame",
			body: `{"message":{"role":"assistant","content":"Ghost"},"done":false}
{"message":{"content":" "},"done":false}
{"message":{"content":"online.\n"},"done":true}`,
			wantChunks: []string{"Ghost", " ", "online.\n"},
		},
		{
			name: "skips empty content without reusing the previous frame",
			body: `{"message":{"content":"Ghost"},"done":false}
{"message":{"content":""},"done":false}
{"done":true}`,
			wantChunks: []string{"Ghost"},
		},
		{
			name: "accepts explicitly completed empty response",
			body: `{"done":true}`,
		},
		{
			name: "stops at completion",
			body: `{"done":true}
{"message":{"content":"obsolete"},"done":false}`,
		},
		{
			name:            "rejects empty stream",
			wantErr:         io.EOF,
			wantErrContains: "decode Ollama chat stream",
		},
		{
			name:            "rejects stream without completion",
			body:            `{"message":{"content":"Partial"},"done":false}`,
			wantChunks:      []string{"Partial"},
			wantErr:         io.EOF,
			wantErrContains: "decode Ollama chat stream",
		},
		{
			name:            "returns malformed frame error",
			body:            `{`,
			wantErr:         io.ErrUnexpectedEOF,
			wantErrContains: "decode Ollama chat stream",
		},
		{
			name: "returns malformed frame error after content",
			body: `{"message":{"content":"Partial"},"done":false}
{`,
			wantChunks:      []string{"Partial"},
			wantErr:         io.ErrUnexpectedEOF,
			wantErrContains: "decode Ollama chat stream",
		},
		{
			name:            "rejects invalid field type",
			body:            `{"done":"true"}`,
			wantErrContains: "decode Ollama chat stream",
		},
		{
			name:            "returns Ollama HTTP error",
			statusCode:      http.StatusNotFound,
			body:            `{"error":"model not found"}`,
			wantErrContains: "404 Not Found: model not found",
		},
		{
			name:            "returns HTTP status for malformed error body",
			statusCode:      http.StatusInternalServerError,
			body:            `unavailable`,
			wantErrContains: "500 Internal Server Error",
		},
		{
			name:            "returns HTTP status for empty error message",
			statusCode:      http.StatusBadRequest,
			body:            `{"error":" "}`,
			wantErrContains: "400 Bad Request",
		},
		{
			name: "returns API error after content with HTTP success",
			body: `{"message":{"content":"Partial"},"done":false}
{"error":"generation failed"}`,
			wantChunks:      []string{"Partial"},
			wantErrContains: "Ollama chat stream failed: generation failed",
		},
		{
			name:            "prioritizes API error over content and completion",
			body:            `{"error":"generation failed","message":{"content":"invalid"},"done":true}`,
			wantErrContains: "Ollama chat stream failed: generation failed",
		},
		{
			name: "stops when content callback fails",
			body: `{"message":{"content":"Ghost"},"done":false}
{"message":{"content":"obsolete"},"done":true}`,
			callbackErr:     callbackErr,
			wantChunks:      []string{"Ghost"},
			wantErr:         callbackErr,
			wantErrContains: "deliver Ollama chat content",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
				if !gotRequest.Stream {
					t.Error("request stream = false, want true")
				}
				wantMessages := []chatMessage{
					{Role: "system", Content: "You are Ghost."},
					{Role: "user", Content: "Identify yourself."},
				}
				if !slices.Equal(gotRequest.Messages, wantMessages) {
					t.Errorf("request messages = %#v, want %#v", gotRequest.Messages, wantMessages)
				}

				writer.Header().Set("Content-Type", "application/x-ndjson")
				if test.statusCode != 0 {
					writer.WriteHeader(test.statusCode)
				}
				_, _ = writer.Write([]byte(test.body))
			}))
			t.Cleanup(server.Close)

			httpClient := server.Client()
			transport := httpClient.Transport
			var body *trackedResponseBody
			httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
				response, err := transport.RoundTrip(request)
				if err != nil {
					return nil, err
				}
				body = &trackedResponseBody{ReadCloser: response.Body}
				response.Body = body
				return response, nil
			})

			client, err := NewClient(server.URL+"/api", httpClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v, want nil", err)
			}

			var gotChunks []string
			err = client.ChatStream(context.Background(), "qwen3:8b", messages, func(content string) error {
				gotChunks = append(gotChunks, content)
				return test.callbackErr
			})

			if body == nil || !body.closed {
				t.Error("ChatStream() response body was not closed")
			}
			if !slices.Equal(gotChunks, test.wantChunks) {
				t.Errorf("ChatStream() chunks = %#v, want %#v", gotChunks, test.wantChunks)
			}
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Errorf("ChatStream() error = %v, want errors.Is(error, %v)", err, test.wantErr)
			}
			if test.wantErrContains != "" {
				if err == nil {
					t.Fatal("ChatStream() error = nil, want an error")
				}
				if !strings.Contains(err.Error(), test.wantErrContains) {
					t.Errorf("ChatStream() error = %q, want it to contain %q", err, test.wantErrContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("ChatStream() error = %v, want nil", err)
			}
		})
	}
}

func TestClientChatStreamTransportError(t *testing.T) {
	transportErr := errors.New("connection unavailable")
	tests := []struct {
		name string
		err  error
	}{
		{name: "preserves transport error", err: transportErr},
		{name: "preserves cancellation", err: context.Canceled},
		{name: "preserves deadline", err: context.DeadlineExceeded},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return nil, test.err
			})}
			client, err := NewClient(DefaultURL, httpClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v, want nil", err)
			}

			err = client.ChatStream(context.Background(), "qwen3:8b", nil, func(_ string) error {
				t.Error("ChatStream() called content callback after transport failure")
				return nil
			})
			if !errors.Is(err, test.err) {
				t.Errorf("ChatStream() error = %v, want errors.Is(error, %v)", err, test.err)
			}
			if err == nil || !strings.Contains(err.Error(), "send Ollama chat request") {
				t.Errorf("ChatStream() error = %v, want send context", err)
			}
		})
	}
}

func TestClientChatStreamReadError(t *testing.T) {
	readErr := errors.New("connection interrupted")
	tests := []struct {
		name       string
		prefix     string
		wantChunks []string
	}{
		{name: "preserves read error before content"},
		{
			name:       "preserves read error after content",
			prefix:     "{\"message\":{\"content\":\"Partial\"},\"done\":false}\n",
			wantChunks: []string{"Partial"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &trackedResponseBody{ReadCloser: io.NopCloser(io.MultiReader(
				strings.NewReader(test.prefix),
				failingReader{err: readErr},
			))}
			httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
			})}
			client, err := NewClient(DefaultURL, httpClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v, want nil", err)
			}

			var gotChunks []string
			err = client.ChatStream(context.Background(), "qwen3:8b", nil, func(content string) error {
				gotChunks = append(gotChunks, content)
				return nil
			})
			if !errors.Is(err, readErr) {
				t.Errorf("ChatStream() error = %v, want errors.Is(error, %v)", err, readErr)
			}
			if err == nil || !strings.Contains(err.Error(), "decode Ollama chat stream") {
				t.Errorf("ChatStream() error = %v, want decode context", err)
			}
			if !slices.Equal(gotChunks, test.wantChunks) {
				t.Errorf("ChatStream() chunks = %#v, want %#v", gotChunks, test.wantChunks)
			}
			if !body.closed {
				t.Error("ChatStream() response body was not closed after read failure")
			}
		})
	}
}

func TestClientChatStreamBufferedCancellation(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		cancelBeforeStream bool
		wantChunks         []string
	}{
		{
			name:               "stops before consuming a cancelled response",
			body:               `{"message":{"content":"obsolete"},"done":true}`,
			cancelBeforeStream: true,
		},
		{
			name: "does not deliver buffered content after callback cancels",
			body: `{"message":{"content":"Partial"},"done":false}
{"message":{"content":"obsolete"},"done":true}`,
			wantChunks: []string{"Partial"},
		},
		{
			name:       "does not report completion when final callback cancels",
			body:       `{"message":{"content":"Final"},"done":true}`,
			wantChunks: []string{"Final"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			body := &trackedResponseBody{ReadCloser: io.NopCloser(strings.NewReader(test.body))}
			httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Context() != ctx {
					t.Error("request context differs from supplied context")
				}
				if test.cancelBeforeStream {
					cancel()
				}
				return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
			})}
			client, err := NewClient(DefaultURL, httpClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v, want nil", err)
			}

			var gotChunks []string
			err = client.ChatStream(ctx, "qwen3:8b", nil, func(content string) error {
				gotChunks = append(gotChunks, content)
				cancel()
				return nil
			})
			if !errors.Is(err, context.Canceled) {
				t.Errorf("ChatStream() error = %v, want context.Canceled", err)
			}
			if !slices.Equal(gotChunks, test.wantChunks) {
				t.Errorf("ChatStream() chunks = %#v, want %#v", gotChunks, test.wantChunks)
			}
			if !body.closed {
				t.Error("ChatStream() response body was not closed after cancellation")
			}
		})
	}
}

func TestClientChatStreamCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-ndjson")
		if err := json.NewEncoder(writer).Encode(chatResponse{Message: chatMessage{Content: "Partial"}}); err != nil {
			t.Errorf("encode response: %v", err)
			return
		}
		writer.(http.Flusher).Flush()
		select {
		case <-request.Context().Done():
		case <-ctx.Done():
		}
	}))
	t.Cleanup(server.Close)

	httpClient := server.Client()
	transport := httpClient.Transport
	var body *trackedResponseBody
	httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		response, err := transport.RoundTrip(request)
		if err != nil {
			return nil, err
		}
		body = &trackedResponseBody{ReadCloser: response.Body}
		response.Body = body
		return response, nil
	})
	client, err := NewClient(server.URL+"/api", httpClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v, want nil", err)
	}

	var gotChunks []string
	err = client.ChatStream(ctx, "qwen3:8b", nil, func(content string) error {
		gotChunks = append(gotChunks, content)
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ChatStream() error = %v, want context.Canceled", err)
	}
	if !slices.Equal(gotChunks, []string{"Partial"}) {
		t.Errorf("ChatStream() chunks = %#v, want %#v", gotChunks, []string{"Partial"})
	}
	if body == nil || !body.closed {
		t.Error("ChatStream() response body was not closed after cancellation")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

type trackedResponseBody struct {
	io.ReadCloser
	closed bool
}

func (body *trackedResponseBody) Close() error {
	body.closed = true
	return body.ReadCloser.Close()
}

type failingReader struct {
	err error
}

func (reader failingReader) Read(_ []byte) (int, error) {
	return 0, reader.err
}

func TestClientChat(t *testing.T) {
	messages := []conversation.Message{
		{
			Role:    conversation.RoleUser,
			Content: "Identify yourself.",
		},
	}

	wantRequestMessage := chatMessage{
		Role:    "user",
		Content: "Identify yourself.",
	}

	responseMessage := chatMessage{
		Role:    "assistant",
		Content: "Ghost online.",
	}

	wantMessage := conversation.Message{
		Role:    conversation.RoleAssistant,
		Content: "Ghost online.",
	}

	tests := []struct {
		name            string
		handler         http.HandlerFunc
		wantMessage     conversation.Message
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

				if len(gotRequest.Messages) != 1 || gotRequest.Messages[0] != wantRequestMessage {
					t.Errorf("request messages = %#v, want %#v", gotRequest.Messages, []chatMessage{wantRequestMessage})
				}

				writer.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(writer).Encode(chatResponse{Message: responseMessage}); err != nil {
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
