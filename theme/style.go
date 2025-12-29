package theme

import (
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
)

var (
	FgAccent0 = lipgloss.NewStyle().Foreground(Accent0)
	FgText    = lipgloss.NewStyle().Foreground(Text)
)

// WordWrap styles and wraps content to width.
func WordWrap(width int, content string) string {
	styled := FgText.Render(content)

	return wordwrap.String(styled, width)
}
