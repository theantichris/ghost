package theme

import (
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
)

var (
	FgAccent0       = lipgloss.NewStyle().Foreground(Accent0)
	FgText          = lipgloss.NewStyle().Foreground(Text)
	JSONKey         = lipgloss.NewStyle().Foreground(Accent1)
	JSONString      = lipgloss.NewStyle().Foreground(SyntaxString)
	JSONNumber      = lipgloss.NewStyle().Foreground(SyntaxString)
	JSONBoolNull    = lipgloss.NewStyle().Foreground(SyntaxType)
	JSONBracket     = lipgloss.NewStyle().Foreground(SyntaxType)
	JSONPunctuation = lipgloss.NewStyle().Foreground(Text)
)

// WordWrap styles and wraps content to width.
func WordWrap(width int, content string, style lipgloss.Style) string {
	styled := style.Render(content)

	return wordwrap.String(styled, width)
}
