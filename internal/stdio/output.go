package stdio

import (
	"io"
	"strings"

	"github.com/charmbracelet/log"
)

const (
	// openTag is the opening delimiter for think blocks in LLM responses.
	openTag string = "<think>"
	// closeTag is the closing delimiter for think blocks in LLM responses.
	closeTag string = "</think>"
)

// OutputWriter handles streaming LLM output with think block filtering.
// It accumulates all tokens while stripping <think>...</think> blocks from
// the output stream. State must be reset between LLM calls using reset().
type OutputWriter struct {
	Logger           *log.Logger
	Output           io.Writer
	Tokens           *string
	buffer           strings.Builder
	insideThinkBlock bool
	canPassThrough   bool
	newlinesTrimmed  bool
}

// Write processes a single token from the LLM stream, accumulating it while
// filtering out think blocks from the output. It trims leading whitespace
// before the first visible output or after closing a think block.
func (writer *OutputWriter) Write(token string) {
	writer.Logger.Debug("token received", "token", token, "length", len(token))

	*writer.Tokens += token

	if writer.canPassThrough {
		output := token

		if !writer.newlinesTrimmed {
			output = strings.TrimLeft(output, " \n\r\t")

			if output != "" {
				writer.newlinesTrimmed = true
			}
		}

		if output != "" {
			writer.Logger.Debug("writing to output", "content", output, "length", len(output))

			if _, err := writer.Output.Write([]byte(output)); err != nil {
				writer.Logger.Error(ErrIO.Error(), "error", err)
			}
		}

		return
	}

	writer.buffer.WriteString(token)
	output := writer.buffer.String()

	if !writer.insideThinkBlock {
		if isOpenTag(output) {
			writer.insideThinkBlock = true

			writer.Logger.Debug("think block opened")
		} else if thinkBlockNotPossible(output) {
			output = strings.TrimLeft(output, " \n\r\t")

			writer.Logger.Debug("writing to output", "content", output, "length", len(output))

			if _, err := writer.Output.Write([]byte(output)); err != nil {
				writer.Logger.Error(ErrIO.Error(), "error", err)
			}

			writer.buffer.Reset()
			writer.canPassThrough = true

			writer.Logger.Debug("pass through mode enabled", "reason", "no think block")
		}
	}

	if writer.insideThinkBlock {
		isCloseTag, index := isCloseTag(output)

		writer.Logger.Debug("checking for close tag", "found", isCloseTag, "index", index, "buffer_len", len(output))

		if isCloseTag {
			output := output[index+len(closeTag):]
			output = strings.TrimLeft(output, " \n\r\t")

			if output != "" {
				writer.Logger.Debug("writing to output", "content", output, "length", len(output))

				if _, err := writer.Output.Write([]byte(output)); err != nil {
					writer.Logger.Error(ErrIO.Error(), "error", err)
				}
			}

			writer.buffer.Reset()
			writer.canPassThrough = true

			writer.Logger.Debug("pass through mode enabled", "reason", "think block closed")
		} else {
			writer.Logger.Debug("buffering think block", "buffer_len", len(output))
		}
	}
}

// Reset clears all outputWriter state for reuse with a new LLM response.
// Reset includes the buffer, state flags, and tokens pointer.
func (writer *OutputWriter) Reset() {
	writer.Logger.Debug("writer reset", "tokensLength", len(*writer.Tokens), "buffLength", writer.buffer.Len(), "insideThinkBlock", writer.insideThinkBlock, "canPassThrough", writer.canPassThrough)

	writer.buffer.Reset()
	writer.insideThinkBlock = false
	writer.canPassThrough = false
	writer.newlinesTrimmed = false
	*writer.Tokens = ""
}

// Flush writes any remaining buffered content to output when the stream ends.
// If inside an unclosed think block, buffered content is discarded as incomplete
// reasoning. If still determining whether content is a think block, the buffer
// is written to output after trimming leading whitespace.
func (writer *OutputWriter) Flush() {
	if writer.buffer.Len() == 0 {
		return
	}

	bufferContent := writer.buffer.String()

	writer.Logger.Debug("flushing buffer", "length", writer.buffer.Len(), "inside_think", writer.insideThinkBlock)

	if writer.insideThinkBlock {
		preview := bufferContent
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		writer.Logger.Debug("discarding incomplete think block", "content_length", len(bufferContent), "preview", preview)
		return
	}

	bufferContent = strings.TrimLeft(bufferContent, " \n\r\t")

	if bufferContent != "" {
		writer.Logger.Debug("writing buffered content", "content", bufferContent, "length", len(bufferContent))

		if _, err := writer.Output.Write([]byte(bufferContent)); err != nil {
			writer.Logger.Error(ErrIO.Error(), "error", err)
		}
	}
}

// isOpenTag checks if content starts with the think block opening tag.
func isOpenTag(content string) bool {
	return strings.HasPrefix(content, openTag)
}

// isCloseTag checks if content contains the think block closing tag and returns
// true and the tag's index if found, false and -1 otherwise.
func isCloseTag(content string) (bool, int) {
	index := strings.Index(content, closeTag)

	return index != -1, index
}

// thinkBlockNotPossible determines if the buffered content cannot possibly
// start a think block by checking if the content is too long without matching
// the opening tag prefix or if it's too short to be a complete opening tag.
func thinkBlockNotPossible(content string) bool {
	return (len(content) >= 7 && !strings.HasPrefix(content, "<think>")) || (len(content) < len(openTag) && !strings.HasPrefix("<think>", content))
}
