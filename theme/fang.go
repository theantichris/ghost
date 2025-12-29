package theme

import (
	"fmt"
	"image/color"
	"io"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
)

var colorScheme = fang.ColorScheme{
	Base:           Text,
	Title:          Accent0,
	Description:    Accent1,
	Codeblock:      Bg1,
	Program:        Accent2,
	DimmedArgument: TextMuted,
	Comment:        SyntaxComment,
	Flag:           TermGreen,
	FlagDefault:    TermBrightYellow,
	Command:        Accent3,
	QuotedString:   SyntaxString,
	Argument:       Accent1,
	Help:           Text,
	Dash:           TextMuted,
	ErrorHeader: [2]color.Color{
		Error,
		Bg1,
	},
	ErrorDetails: TermBrightRed,
}

// GetFangColorScheme is called by Fang to get the color scheme.
func GetFangColorScheme(ld lipgloss.LightDarkFunc) fang.ColorScheme {
	return colorScheme
}

// FangErrorHandler renders error messages with styles.
func FangErrorHandler(w io.Writer, styles fang.Styles, err error) {
	headerStyle := lipgloss.NewStyle().
		Foreground(Error)

	messageStyle := lipgloss.NewStyle().
		Foreground(Error)

	header := headerStyle.Render("󱙜 error")
	message := messageStyle.Render(err.Error())

	fmt.Fprintf(w, "%s: %s\n", header, message)
}
