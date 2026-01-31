package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/go-cmp/cmp"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func TestEncodeImage(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		err         error
	}{
		{
			name:        "encodes file content to base64",
			fileContent: "test image content",
		},
		{
			name:        "encodes binary content",
			fileContent: "\x89PNG\r\n\x1a\n",
		},
		{
			name:        "encodes empty file",
			fileContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "ghost-test-*.png")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			defer func() {
				_ = os.Remove(tmpFile.Name())
			}()

			_, err = tmpFile.WriteString(tt.fileContent)
			if err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}

			_ = tmpFile.Close()

			got, err := encodeImage(tmpFile.Name())

			if tt.wantErr {
				if err == nil {
					t.Fatal("encodeImage() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("encodeImage() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("encodeImage() err = %v, want nil", err)
			}

			want := base64.StdEncoding.EncodeToString([]byte(tt.fileContent))
			if got != want {
				t.Errorf("encodeImage() = %q, want %q", got, want)
			}
		})
	}
}

func TestEncodeImage_NonExistentFile(t *testing.T) {
	_, err := encodeImage("/nonexistent/path/to/image.png")

	if err == nil {
		t.Fatal("encodeImage() err = nil, want error")
	}

	if !errors.Is(err, ErrImageAnalysis) {
		t.Errorf("encodeImage() err = %v, want %v", err, ErrImageAnalysis)
	}
}

func TestInitMessages(t *testing.T) {
	tests := []struct {
		name   string
		system string
		prompt string
		format string
		want   []llm.ChatMessage
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			prompt: "user prompt",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			prompt: "user prompt",
			format: "json",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: jsonPrompt},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with markdown format",
			system: "system prompt",
			prompt: "user prompt",
			format: "markdown",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: markdownPrompt},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "ignores unknown format",
			system: "system prompt",
			prompt: "user prompt",
			format: "unknown",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := initMessages(tt.system, tt.prompt, tt.format)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("initMessages() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAnalyseImages(t *testing.T) {
	tests := []struct {
		name            string
		images          []string
		mockResponses   []string
		mockStatusCodes []int
		wantMsgCount    int
		wantErr         bool
		err             error
	}{
		{
			name:         "returns empty slice when no images provided",
			images:       []string{},
			wantMsgCount: 0,
		},
		{
			name:   "analyzes single image",
			images: []string{"PLACEHOLDER"},
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"Image analysis result"}}`,
			},
			mockStatusCodes: []int{http.StatusOK},
			wantMsgCount:    1,
		},
		{
			name:   "analyzes multiple images",
			images: []string{"PLACEHOLDER1", "PLACEHOLDER2"},
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"Analysis 1"}}`,
				`{"message":{"role":"assistant","content":"Analysis 2"}}`,
			},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			wantMsgCount:    2,
		},
		{
			name:   "returns error when API fails",
			images: []string{"PLACEHOLDER"},
			mockResponses: []string{
				`{"error":"internal error"}`,
			},
			mockStatusCodes: []int{http.StatusInternalServerError},
			wantErr:         true,
			err:             llm.ErrUnexpectedStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idx := int(callCount.Add(1)) - 1

				if idx >= len(tt.mockResponses) {
					t.Errorf("unexpected request #%d", idx+1)
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				w.WriteHeader(tt.mockStatusCodes[idx])
				_, _ = w.Write([]byte(tt.mockResponses[idx]))
			}))
			defer server.Close()

			logger := log.New(io.Discard)

			// Create temp files for images that need them.
			var imagePaths []string
			for _, img := range tt.images {
				if img == "" {
					continue
				}

				tmpFile, err := os.CreateTemp("", "ghost-test-*.png")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				defer func(name string) {
					_ = os.Remove(name)
				}(tmpFile.Name())

				_, _ = tmpFile.WriteString("fake image data")
				_ = tmpFile.Close()

				imagePaths = append(imagePaths, tmpFile.Name())
			}

			got, err := AnalyseImages(context.Background(), server.URL, "test-model", imagePaths, logger)

			if tt.wantErr {
				if err == nil {
					t.Fatal("AnalyseImages() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("AnalyseImages() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("AnalyseImages() err = %v, want nil", err)
			}

			if len(got) != tt.wantMsgCount {
				t.Errorf("AnalyseImages() message count = %d, want %d", len(got), tt.wantMsgCount)
			}

			// Verify all returned messages are user role with content.
			for i, msg := range got {
				if msg.Role != llm.RoleUser {
					t.Errorf("AnalyseImages() message[%d].Role = %v, want %v", i, msg.Role, llm.RoleUser)
				}

				if msg.Content == "" {
					t.Errorf("AnalyseImages() message[%d].Content is empty", i)
				}
			}
		})
	}
}

func TestAnalyseImages_FileReadError(t *testing.T) {
	logger := log.New(io.Discard)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected API call")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := AnalyseImages(context.Background(), server.URL, "test-model", []string{"/nonexistent/image.png"}, logger)

	if err == nil {
		t.Fatal("AnalyseImages() err = nil, want error")
	}

	if !errors.Is(err, ErrImageAnalysis) {
		t.Errorf("AnalyseImages() err = %v, want %v", err, ErrImageAnalysis)
	}
}
