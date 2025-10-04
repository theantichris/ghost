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
			if _, err := writer.output.Write([]byte(output)); err != nil {
				writer.logger.Error(ErrIO.Error(), "error", err)
			}
		}

		return
	}

	writer.buffer.WriteString(token)
	bufferContent := writer.buffer.String()

	if !writer.insideThinkBlock {
		if isOpenTag(bufferContent) {
			writer.insideThinkBlock = true
		} else if thinkBlockNotPossible(bufferContent) {
			bufferContent = strings.TrimLeft(bufferContent, " \n\r\t")

			if _, err := writer.output.Write([]byte(bufferContent)); err != nil {
				writer.logger.Error("%w: %w", ErrIO, err)
			}

			writer.buffer.Reset()
			writer.canPassThrough = true
		}
	}

	if writer.insideThinkBlock {
		isCloseTag, index := isCloseTag(bufferContent)

		if isCloseTag {
			output := bufferContent[index+len(closeTag):]
			output = strings.TrimLeft(output, " \n\r\t")

			if output != "" {
				if _, err := writer.output.Write([]byte(output)); err != nil {
					writer.logger.Error("%w: %w", ErrIO, err)
				}
			}

			writer.buffer.Reset()
			writer.canPassThrough = true
		}
	}
}

// reset clears all outputWriter state for reuse with a new LLM response.
// This includes the buffer, state flags, and tokens pointer.
func (writer *outputWriter) reset() {
	writer.buffer.Reset()
	writer.insideThinkBlock = false
	writer.canPassThrough = false
	writer.newlinesTrimmed = false
	*writer.tokens = ""
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
