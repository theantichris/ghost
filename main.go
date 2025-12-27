package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/carlmjohnson/requests"
)

const (
	host   = "http://localhost:11434/api"
	model  = "dolphin-mixtral:8x7b"
	system = "You are ghost, a cyberpunk AI assistant."
)

var (
	errPromptNotDetected = errors.New("prompt not detected")
)

// chatRequest holds the information for the chat endpoint.
type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

// chatMessage holds a single message in the chat history.
type chatMessage struct {
	// Role holds the author of the message.
	// Values are system, user, assistant, tool.
	Role string `json:"role"`

	// Content holds the message history.
	Content string `json:"content"`
}

// chatResponse holds the response from the chat endpoint.
type chatResponse struct {
	Message chatMessage `json:"message"`
}

func main() {
	prompt, err := getPrompt(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	messages := createMessages(system, prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	chatResponse, err := getChatResponse(ctx, model, messages)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(chatResponse)
}

func getPrompt(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("%w", errPromptNotDetected)
	}

	return args[1], nil
}

func createMessages(system, prompt string) []chatMessage {
	messages := []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: prompt},
	}

	return messages
}

func getChatResponse(ctx context.Context, model string, messages []chatMessage) (string, error) {
	request := chatRequest{
		Model:    model,
		Stream:   false,
		Messages: messages,
	}

	var chatResponse chatResponse

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		ToJSON(&chatResponse).
		Fetch(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return chatResponse.Message.Content, nil
}
