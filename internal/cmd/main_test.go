package cmd

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"

	"github.com/urfave/cli/v3"
)

func TestBefore(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "initializes LLM Client and adds to metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			cmd := cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Value: "http://test.dev",
					},
					&cli.StringFlag{
						Name:  "model",
						Value: "default:model",
					},
					&cli.StringFlag{
						Name:  "vision-model",
						Value: "vision:model",
					},
				},
				Metadata: map[string]any{
					"logger": logger,
				},
			}

			_, err := beforeHook(context.Background(), &cmd)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			actual := cmd.Metadata["llmClient"]

			if actual == nil {
				t.Fatalf("expected Client, got nil")
			}

			if _, ok := actual.(llm.Client); !ok {
				t.Errorf("expected LLM Client to be of type Client, got %v", actual)
			}
		})
	}

}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		images       []string
		streamChunks []string
		wantCalls    []string
		config       config
		llmClient    llm.Client
		expected     string
		wantErr      bool
		Error        error
	}{
		{
			name:   "sends prompt to LLM generate without images",
			prompt: "test prompt",
			images: []string{},
			config: config{
				systemPrompt: "system prompt",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					return callback("this prompt is good")
				},
			},
			expected: "this prompt is good",
		},
		{
			name:   "returns error for LLM generate without images",
			prompt: "test prompt",
			images: []string{},
			config: config{
				systemPrompt: "system prompt",
			},
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
			wantErr: true,
			Error:   llm.ErrOllama,
		},
		{
			name:   "sends prompt to LLM generate with images",
			prompt: "test prompt",
			images: []string{"test/image.png"},
			config: config{
				systemPrompt:       "system prompt",
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					return callback("this prompt is good")
				},
			},
			expected: "this prompt is good",
		},
		{
			name:   "returns error for LLM generate with images",
			prompt: "test prompt",
			images: []string{"test/image.png"},
			config: config{
				systemPrompt:       "system prompt",
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
			wantErr: true,
			Error:   llm.ErrOllama,
		},
		{
			name:         "streams text chunks to callback",
			prompt:       "test prompt",
			images:       nil,
			streamChunks: []string{"Hello", " ", "World", "!"},
			config: config{
				systemPrompt: "system prompt",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					for _, chunk := range []string{"Hello", " ", "World", "!"} {
						if err := callback(chunk); err != nil {
							return err
						}
					}
					return nil
				},
			},
			wantCalls: []string{"Hello", " ", "World", "!"},
			expected:  "Hello World!",
		},
		{
			name:   "handles vision model then chat model streaming",
			prompt: "describe this",
			images: []string{"image1"},
			config: config{
				systemPrompt:       "system prompt",
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func() func(context.Context, string, string, string, []string, func(string) error) error {
					callCount := 0
					return func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
						if callCount == 0 {
							// First call: vision model
							callCount++
							return callback("Image analysis")
						}

						// Second call: chat model
						return callback("Final response")
					}
				}(),
			},
			wantCalls: []string{"Final response"},
			expected:  "Final response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string

			streamCallback := func(chunk string) error {
				calls = append(calls, chunk)

				return nil
			}

			err := generate(context.Background(), tt.prompt, tt.images, tt.config, tt.llmClient, streamCallback)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if tt.wantCalls != nil {
					if len(calls) != len(tt.wantCalls) {
						t.Errorf("expected %d calls, got %d", len(tt.wantCalls), len(calls))
					}

					for i, want := range tt.wantCalls {
						if i >= len(calls) {
							break
						}

						if calls[i] != want {
							t.Errorf("callbaack call %d: expected %q, got %q", i, want, calls[i])
						}
					}
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if !errors.Is(err, tt.Error) {
					t.Errorf("expected error %v, got %v", tt.Error, err)
				}
			}
		})
	}
}

func TestAnalyzeImages(t *testing.T) {
	tests := []struct {
		name        string
		images      []string
		config      config
		llmClient   llm.Client
		expected    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "returns vision analysis",
			images: []string{"image1.jpg"},
			config: config{
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "analyze this",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					return callback("This is a cat")
				},
			},
			expected: "This is a cat",
		},
		{
			name:   "accumulates multiple chunks",
			images: []string{"image1.png"},
			config: config{
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					chunks := []string{"This ", "is ", "a ", "cat"}

					for _, chunk := range chunks {
						if err := callback(chunk); err != nil {
							return err
						}
					}

					return nil
				},
			},
			expected: "This is a cat",
		},
		{
			name:   "returns error from LLM",
			images: []string{"image1.png"},
			config: config{
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
			wantErr:     true,
			expectedErr: llm.ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzeImages(context.Background(), tt.llmClient, tt.config, tt.images)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateResponse(t *testing.T) {
	tests := []struct {
		name         string
		config       config
		systemPrompt string
		prompt       string
		llmClient    llm.Client
		wantCalls    []string
		expected     string
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "generates and streams response",
			config: config{
				model:        "test model",
				systemPrompt: "system prompt",
			},
			prompt: "user prompt",
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					return callback("Hello world")
				},
			},
			wantCalls: []string{"Hello world"},
			expected:  "Hello world",
		},
		{
			name: "accumulates and streams multiple chunks",
			config: config{
				model:        "test model",
				systemPrompt: "system prompt",
			},
			prompt: "user prompt",
			llmClient: llm.MockClient{
				GenerateFunc: func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
					chunks := []string{"Hello", " ", "world", "!"}
					for _, chunk := range chunks {
						if err := callback(chunk); err != nil {
							return err
						}
					}
					return nil
				},
			},
			wantCalls: []string{"Hello", " ", "world", "!"},
			expected:  "Hello world!",
		},
		{
			name: "returns error from LLM",
			config: config{
				model:        "test model",
				systemPrompt: "system prompt",
			},
			prompt: "user prompt",
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
			wantErr:     true,
			expectedErr: llm.ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			streamCallback := func(chunk string) error {
				calls = append(calls, chunk)
				return nil
			}

			err := generateResponse(context.Background(), tt.llmClient, tt.config, tt.prompt, streamCallback)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.wantCalls != nil {
				if len(calls) != len(tt.wantCalls) {
					t.Errorf("expected %d callback calls, got %d", len(tt.wantCalls), len(calls))
				}
				for i, want := range tt.wantCalls {
					if i >= len(calls) {
						break
					}
					if calls[i] != want {
						t.Errorf("callback call %d: expected %q, got %q", i, want, calls[i])
					}
				}
			}
		})
	}
}
