package cmd

import (
	"io"
	"strings"

	"github.com/charmbracelet/log"
)

const (
	openTag  string = "<think>"
	closeTag string = "</think>"
)

// outputWriter handles streaming LLM output with think block filtering.
// It accumulates all tokens while stripping <think>...</think> blocks from
// the output stream. State must be reset between LLM calls using reset().
type outputWriter struct {
	logger           *log.Logger
	output           io.Writer
	tokens           *string
	buffer           strings.Builder
	insideThinkBlock bool
	canPassThrough   bool
	newlinesTrimmed  bool
}

// write processes a single token from the LLM stream, accumulating it while
// filtering out think blocks from the output. It trims leading whitespace
// before the first visible output or after closing a think block.
func (writer *outputWriter) write(token string) {
	writer.logger.Debug("token received", "token", token, "length", len(token))

	*writer.tokens += token

	if writer.canPassThrough {
		output := token

		if !writer.newlinesTrimmed {
			output = strings.TrimLeft(output, " \n\r\t")

			if output != "" {
				writer.newlinesTrimmed = true
			}
		}

		if output != "" {
			writer.logger.Debug("writing to output", "content", output, "length", len(output))

			if _, err := writer.output.Write([]byte(output)); err != nil {
				writer.logger.Error(ErrIO.Error(), "error", err)
			}
		}

		return
	}

	writer.buffer.WriteString(token)
	output := writer.buffer.String()

	if !writer.insideThinkBlock {
		if isOpenTag(output) {
			writer.insideThinkBlock = true

			writer.logger.Debug("think block opened")
		} else if thinkBlockNotPossible(output) {
			output = strings.TrimLeft(output, " \n\r\t")

			writer.logger.Debug("writing to output", "content", output, "length", len(output))

			if _, err := writer.output.Write([]byte(output)); err != nil {
				writer.logger.Error("%w: %w", ErrIO, err)
			}

			writer.buffer.Reset()
			writer.canPassThrough = true

			writer.logger.Debug("pass through mode enabled", "reason", "no think block")
		}
	}

	if writer.insideThinkBlock {
		isCloseTag, index := isCloseTag(output)

		if isCloseTag {
			output := output[index+len(closeTag):]
			output = strings.TrimLeft(output, " \n\r\t")

			if output != "" {
				writer.logger.Debug("writing to output", "content", output, "length", len(output))

				if _, err := writer.output.Write([]byte(output)); err != nil {
					writer.logger.Error("%w: %w", ErrIO, err)
				}
			}

			writer.buffer.Reset()
			writer.canPassThrough = true

			writer.logger.Debug("pass through mode enabled", "reason", "think block closed")
		}
	}
}

// reset clears all outputWriter state for reuse with a new LLM response.
// This includes the buffer, state flags, and tokens pointer.
func (writer *outputWriter) reset() {
	writer.logger.Debug("writer reset", "tokensLength", len(*writer.tokens), "buffLength", writer.buffer.Len(), "insideThinkBlock", writer.insideThinkBlock, "canPassThrough", writer.canPassThrough)

	writer.buffer.Reset()
	writer.insideThinkBlock = false
	writer.canPassThrough = false
	writer.newlinesTrimmed = false
	*writer.tokens = ""
}

// flush writes any remaining buffered content to output when the stream ends.
// If inside an unclosed think block, buffered content is discarded as incomplete
// reasoning. If still determining whether content is a think block, the buffer
// is written to output after trimming leading whitespace.
func (writer *outputWriter) flush() {
	if writer.buffer.Len() == 0 {
		return
	}

	bufferContent := writer.buffer.String()

	writer.logger.Debug("flushing buffer", "length", writer.buffer.Len(), "inside_think", writer.insideThinkBlock)

	if writer.insideThinkBlock {
		writer.logger.Debug("discarding incomplete think block", "content_length", len(bufferContent))
		return
	}

	bufferContent = strings.TrimLeft(bufferContent, " \n\r\t")

	if bufferContent != "" {
		writer.logger.Debug("writing buffered content", "content", bufferContent, "length", len(bufferContent))

		if _, err := writer.output.Write([]byte(bufferContent)); err != nil {
			writer.logger.Error(ErrIO.Error(), "error", err)
		}
	}
}

// isOpenTag checks if content starts with the think block opening tag.
func isOpenTag(content string) bool {
	return strings.HasPrefix(content, openTag)
}

// isCloseTag checks if content contains the think block closing tag.
// Returns true and the tag's index if found, false and -1 otherwise.
func isCloseTag(content string) (bool, int) {
	index := strings.Index(content, closeTag)

	return index != -1, index
}

// thinkBlockNotPossible determines if the buffered content cannot possibly
// start a think block, enabling pass-through mode for better performance.
func thinkBlockNotPossible(content string) bool {
	return (len(content) >= 7 && !strings.HasPrefix(content, "<think>")) || (len(content) < len(openTag) && !strings.HasPrefix("<think>", content))
}
