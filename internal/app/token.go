package app

import (
	"io"
	"strings"
)

const (
	openTag  string = "<think>"
	closeTag string = "</think>"
)

// tokenHandler handles the response tokens from the LLM.
type tokenHandler struct {
	output      io.Writer       // The output writer.
	tokens      *string         // Tokens to be output.
	insideThink bool            // True if inside a <think> block.
	buffer      strings.Builder // Buffers the tokens to check for <think> blocks.
	passThrough bool            // True if no possible <think> blocks.
	hasOutput   bool            // Used to trim newlines from the beginning of output.
}

// handle updates the tokens then filters out the think and writes the output.
func (handler *tokenHandler) handle(token string) {
	*handler.tokens += token

	if handler.passThrough {
		outputContent := token

		if !handler.hasOutput {
			outputContent = strings.TrimLeft(outputContent, " \n\r\t")

			if outputContent != "" {
				handler.hasOutput = true
			}
		}

		if outputContent != "" {
			handler.output.Write([]byte(outputContent))
		}

		return
	}

	handler.buffer.WriteString(token)
	bufferContent := handler.buffer.String()

	if !handler.insideThink {
		if strings.HasPrefix(bufferContent, openTag) {
			handler.insideThink = true
		} else if noThinkBlocks(bufferContent) {
			bufferContent = strings.TrimLeft(bufferContent, " \n\r\t")
			handler.output.Write([]byte(bufferContent))

			handler.buffer.Reset()
			handler.passThrough = true
		}
	}

	if handler.insideThink {
		if idx := strings.Index(bufferContent, closeTag); idx != -1 {
			content := bufferContent[idx+len(closeTag):]
			content = strings.TrimLeft(content, " \n\r\t")

			if content != "" {
				handler.output.Write([]byte(content))
			}

			handler.buffer.Reset()
			handler.passThrough = true
		}
	}
}

// noThinkBlocks checks the bufferContent if a <think> block is possible.
func noThinkBlocks(bufferContent string) bool {
	return (len(bufferContent) >= 7 && !strings.HasPrefix(bufferContent, "<think>")) || (len(bufferContent) < len(openTag) && !strings.HasPrefix("<think>", bufferContent))
}
