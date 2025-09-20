package app

import (
	"io"
	"strings"
)

// tokenHandler handles the response tokens from the LLM.
type tokenHandler struct {
	output io.Writer
	tokens *string
}

// handle filters out the <think> block, updates the tokens, and writes the output.
func (handler *tokenHandler) handle(token string) {
	*handler.tokens += token

	openTag := "<think>"
	closeTag := "</think>"

	if strings.HasPrefix(token, openTag) {
		closeIndex := strings.Index(token, closeTag)
		if closeIndex != -1 {
			output := token[closeIndex+len(closeTag):]
			handler.output.Write([]byte(output))
		}
	} else {
		handler.output.Write([]byte(token))
	}
}
