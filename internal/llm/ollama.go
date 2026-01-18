package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
)

var (
	ErrModelNotFound    = errors.New("AI construct not found in the system")
	ErrUnexpectedStatus = errors.New("unexpected response from neural network")
	ErrDecodeChunk      = errors.New("data packet decode error")
)

// Role represents the author of a message in the chat history.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatRequest holds the information for the chat endpoint.
type ChatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []ChatMessage `json:"messages"`
	Tools    []Tool        `json:"tools,omitempty"`
}

// ChatResponse holds the response from the chat endpoint.
type ChatResponse struct {
	Message ChatMessage `json:"message"`
}

// ChatMessage holds a single message in the chat history.
type ChatMessage struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	Images    []string   `json:"images,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// AnalyzeImages sends a request to the chat endpoint with images to analyze and
// returns the response message.
func AnalyzeImages(ctx context.Context, host, model string, messages []ChatMessage) (ChatMessage, error) {
	request := ChatRequest{
		Model:    model,
		Stream:   false,
		Messages: messages,
	}

	var chatResponse ChatResponse

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		AddValidator(nil).
		Handle(func(response *http.Response) error {
			if response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("%w: %s", ErrModelNotFound, request.Model)
			}

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("%w: %s", ErrUnexpectedStatus, response.Status)
			}

			return nil
		}).
		ToJSON(&chatResponse).
		Fetch(ctx)

	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w", err)
	}

	chatMessage := ChatMessage{
		Role:    RoleAssistant,
		Content: chatResponse.Message.Content,
	}

	return chatMessage, nil
}

// Chat sends a non-streaming request to the chat endpoint with tools and returns
// the response message.
func Chat(ctx context.Context, host, model string, messages []ChatMessage, tools []Tool) (ChatMessage, error) {
	request := ChatRequest{
		Model:    model,
		Stream:   false,
		Messages: messages,
		Tools:    tools,
	}

	var chatResponse ChatResponse

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		AddValidator(nil).
		Handle(func(resp *http.Response) error {
			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("%w: %s", ErrModelNotFound, request.Model)
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%w: %s", ErrUnexpectedStatus, resp.Status)
			}

			return nil
		}).
		ToJSON(&chatResponse).
		Fetch(ctx)

	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w", err)
	}

	// Return chatResponse.Message directly to preserve ToolCalls.
	return chatResponse.Message, nil
}

// StreamChat sends a streaming request to the chat endpoint and returns the
// response message.
// onChunk is called for each streamed chunk of content.
func StreamChat(ctx context.Context, host, model string, messages []ChatMessage, tools []Tool, onChunk func(string)) (ChatMessage, error) {
	request := ChatRequest{
		Model:    model,
		Stream:   true,
		Messages: messages,
		Tools:    tools,
	}

	var chatContent strings.Builder

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		AddValidator(nil).
		Handle(func(response *http.Response) error {
			defer func() {
				_ = response.Body.Close()
			}()

			if response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("%w: %s", ErrModelNotFound, request.Model)
			}

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("%w: %s", ErrUnexpectedStatus, response.Status)
			}

			decoder := json.NewDecoder(response.Body)

			for {
				var chunk ChatResponse

				if err := decoder.Decode(&chunk); err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("%w: %w", ErrDecodeChunk, err)
				}

				onChunk(chunk.Message.Content)

				chatContent.WriteString(chunk.Message.Content)
			}

			return nil
		}).
		Fetch(ctx)

	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w", err)
	}

	chatMessage := ChatMessage{
		Role:    RoleAssistant,
		Content: chatContent.String(),
	}

	return chatMessage, nil
}
