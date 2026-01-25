package agent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/tool"
)

func TestRunToolLoop(t *testing.T) {
	tests := []struct {
		name            string
		registerTool    bool
		mockResponses   []string
		mockStatusCodes []int
		toolResult      string
		toolErr         error
		wantMsgCount    int
		wantErr         bool
		err             error
	}{
		{
			name:         "returns messages unchanged when no tools registered",
			registerTool: false,
			wantMsgCount: 1,
		},
		{
			name:         "returns messages when LLM returns no tool calls",
			registerTool: true,
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"Hello!"}}`,
			},
			mockStatusCodes: []int{http.StatusOK},
			wantMsgCount:    1,
		},
		{
			name:         "executes single tool call",
			registerTool: true,
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"mock_tool","arguments":{}}}]}}`,
				`{"message":{"role":"assistant","content":"Done!"}}`,
			},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			toolResult:      "tool result",
			wantMsgCount:    3, // original + assistant with tool call + tool result
		},
		{
			name:         "executes multiple tool calls in single response",
			registerTool: true,
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"mock_tool","arguments":{}}},{"function":{"name":"mock_tool","arguments":{}}}]}}`,
				`{"message":{"role":"assistant","content":"Done!"}}`,
			},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			toolResult:      "tool result",
			wantMsgCount:    4, // original + assistant with tool calls + 2 tool results
		},
		{
			name:         "executes multi-iteration tool loop",
			registerTool: true,
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"mock_tool","arguments":{}}}]}}`,
				`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"mock_tool","arguments":{}}}]}}`,
				`{"message":{"role":"assistant","content":"Done!"}}`,
			},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK, http.StatusOK},
			toolResult:      "tool result",
			wantMsgCount:    5, // original + (assistant + tool result) * 2
		},
		{
			name:         "returns error when LLM request fails",
			registerTool: true,
			mockResponses: []string{
				`{"error":"internal error"}`,
			},
			mockStatusCodes: []int{http.StatusInternalServerError},
			wantMsgCount:    1,
			wantErr:         true,
			err:             llm.ErrUnexpectedStatus,
		},
		{
			name:         "continues loop when tool execution fails",
			registerTool: true,
			mockResponses: []string{
				`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"mock_tool","arguments":{}}}]}}`,
				`{"message":{"role":"assistant","content":"Done!"}}`,
			},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			toolErr:         errors.New("tool failed"),
			wantMsgCount:    3, // original + assistant with tool call + error message
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
			registry := tool.NewRegistry()

			if tt.registerTool {
				registry.Register(tool.MockTool{
					Name:   "mock_tool",
					Result: tt.toolResult,
					Err:    tt.toolErr,
				})
			}

			messages := []llm.ChatMessage{
				{Role: llm.RoleUser, Content: "test"},
			}

			got, err := RunToolLoop(context.Background(), registry, server.URL, "test-model", messages, logger)

			if tt.wantErr {
				if err == nil {
					t.Fatal("RunToolLoop() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("RunToolLoop() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("RunToolLoop() err = %v, want nil", err)
			}

			if len(got) != tt.wantMsgCount {
				t.Errorf("RunToolLoop() message count = %d, want %d", len(got), tt.wantMsgCount)
			}
		})
	}
}
