package theme

import (
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
)

var FgAccent0 = lipgloss.NewStyle().Foreground(Accent0)

// WordWrap wraps content to width.
func WordWrap(width int, content string) string {
	return wordwrap.String(content, width)
}
