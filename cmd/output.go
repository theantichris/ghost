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

type outputWriter struct {
	logger           *log.Logger
	output           io.Writer
	tokens           *string
	buffer           strings.Builder
	insideThinkBlock bool
	canPassThrough   bool
	newlinesTrimmed  bool
}

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

func isOpenTag(content string) bool {
	return strings.HasPrefix(content, openTag)
}

func isCloseTag(content string) (bool, int) {
	index := strings.Index(content, closeTag)

	return index != -1, index
}

func thinkBlockNotPossible(content string) bool {
	return (len(content) >= 7 && !strings.HasPrefix(content, "<think>")) || (len(content) < len(openTag) && !strings.HasPrefix("<think>", content))
}
