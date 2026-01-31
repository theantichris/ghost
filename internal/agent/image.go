package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
)

var ErrImageAnalysis = errors.New("visual recon failed")

// AnalyzeImages sends requests to the LLM to analyze images and returns a slice
// of llm.ChatMessage with the reports.
func AnalyseImages(ctx context.Context, url, visionModel string, images []string, logger *log.Logger) ([]llm.ChatMessage, error) {
	var imageAnalysis []llm.ChatMessage

	// Loop through each image and send it to the LLM for analysis, attact response to returned messages.
	for _, image := range images {
		filename := filepath.Base(image)
		logger.Debug("digitizing visual data", "filename", filename)

		encodedImage, err := encodeImage(image)
		if err != nil {
			return []llm.ChatMessage{}, err
		}

		prompt := fmt.Sprintf("Filename: %s\n\n%s", filename, visionPrompt)
		messages := initMessages(visionSystemPrompt, prompt, "markdown")
		messages[len(messages)-1].Images = []string{encodedImage} // Attach images to user prompt message.

		logger.Info("initializing visual recon", "model", visionModel, "url", url, "filename", filename, "format", "markdown")

		response, err := llm.AnalyzeImages(ctx, url, visionModel, messages)
		if err != nil {
			return []llm.ChatMessage{}, err
		}

		logger.Debug("visual recon complete", "filename", filename)

		imageAnalysis = append(imageAnalysis, llm.ChatMessage{Role: llm.RoleUser, Content: response.Content})
	}

	return imageAnalysis, nil
}

// encodedImage takes an image path and returns a base64 encoded string.
func encodeImage(image string) (string, error) {
	imageBytes, err := os.ReadFile(image)
	if err != nil {
		return "", fmt.Errorf("%w: failed to read %s: %w", ErrImageAnalysis, image, err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)

	return encodedImage, nil
}

// initMessages creates and returns an initial message history.
// TODO: duplicated with cmd/root.go
func initMessages(system, prompt, format string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: system},
	}

	if format != "" {
		switch format {
		case "json":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: JSONPrompt})
		case "markdown":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: MarkdownPrompt})
		}
	}

	messages = append(messages, llm.ChatMessage{Role: llm.RoleUser, Content: prompt})

	return messages
}
