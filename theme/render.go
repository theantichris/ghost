package theme

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/glamour"
)

var ErrMarkdownRender = errors.New("markdown render error")

// RenderContent returns content rendered for the proper format and output.
func RenderContent(content, format string, isTTY bool) (string, error) {
	if !isTTY {
		return content, nil
	}

	switch format {
	case "json":
		return JSON(content), nil

	case "markdown":
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStyles(CyberpunkTheme()),
		)

		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrMarkdownRender, err)
		}

		render, err := renderer.Render(content)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrMarkdownRender, err)
		}

		return render, nil

	default:
		return content, nil
	}
}
