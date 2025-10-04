package cmd

import (
	"io"
	"strings"

	"github.com/charmbracelet/log"
)

type outputWriter struct {
	logger *log.Logger
	output io.Writer
	tokens *string
}

func (outputWriter *outputWriter) writeLLMOutput(token string) {
	*outputWriter.tokens += token

	_, err := outputWriter.output.Write([]byte(token))
	if err != nil {
		outputWriter.logger.Error(ErrIO.Error(), "error", err)
	}
}

// stripThinkBlock removes <think>...</think> blocks from the message.
// These blocks may contain internal reasoning that shouldn't be shown to the user.
func stripThinkBlock(message string) string {
	openTag := "<think>"
	closeTag := "</think>"

	for {
		start := strings.Index(message, openTag)

		if start == -1 {
			break // No <think> block.
		}

		end := strings.Index(message[start+len(openTag):], closeTag)
		if end == -1 {
			break // No </think> block.
		}

		blockEnd := start + len(openTag) + end + len(closeTag)

		message = message[:start] + message[blockEnd:]
	}

	return strings.TrimSpace(message)
}
