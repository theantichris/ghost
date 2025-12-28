package main

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
)

// Cyberpunk Color Palette
// https://github.com/theantichris/dotfiles/blob/main/color_palettes/cyberpunk/palette.html
var (
	Bg0 = lipgloss.Color("#16161E")
	Bg1 = lipgloss.Color("#1F1F28")
	Bg2 = lipgloss.Color("#21222C")
	Bg3 = lipgloss.Color("#2A2A33")

	Text      = lipgloss.Color("#EAEAF2")
	TextMuted = lipgloss.Color("#8A86A0")

	Accent0 = lipgloss.Color("#CA0174")
	Accent1 = lipgloss.Color("#00F0FF")
	Accent2 = lipgloss.Color("#FF00FF")
	Accent3 = lipgloss.Color("#00FFA2")

	Success    = lipgloss.Color("#00FF9C")
	Warning    = lipgloss.Color("#FFD300")
	ErrorColor = lipgloss.Color("#FF003C")
	Info       = lipgloss.Color("#009CFF")
	Link       = lipgloss.Color("#00FFFF")

	TermBlack   = lipgloss.Color("#16161E")
	TermRed     = lipgloss.Color("#FF003C")
	TermGreen   = lipgloss.Color("#00FF9C")
	TermYellow  = lipgloss.Color("#FFD300")
	TermBlue    = lipgloss.Color("#009CFF")
	TermMagenta = lipgloss.Color("#FF00FF")
	TermCyan    = lipgloss.Color("#00F0FF")
	TermWhite   = lipgloss.Color("#EAEAF2")

	TermBrightBlack   = lipgloss.Color("#2A2A33")
	TermBrightRed     = lipgloss.Color("#FF3369")
	TermBrightGreen   = lipgloss.Color("#66FFBF")
	TermBrightYellow  = lipgloss.Color("#FFE766")
	TermBrightBlue    = lipgloss.Color("#33CFFF")
	TermBrightMagenta = lipgloss.Color("#FF66FF")
	TermBrightCyan    = lipgloss.Color("#66F5FF")
	TermBrightWhite   = lipgloss.Color("#FFFFFF")

	SyntaxKeyword  = lipgloss.Color("#CA0174")
	SyntaxString   = lipgloss.Color("#00FFA2")
	SyntaxFunction = lipgloss.Color("#00F0FF")
	SyntaxVariable = lipgloss.Color("#EAEAF2")
	SyntaxType     = lipgloss.Color("#FF00FF")
	SyntaxOperator = lipgloss.Color("#FFD300")
	SyntaxNumber   = lipgloss.Color("#FF7B00")
	SyntaxComment  = lipgloss.Color("#8A86A0")
)

var colorScheme = fang.ColorScheme{
	Base:           Text,             // Main text color
	Title:          Accent0,          // Hot pink for titles
	Description:    Accent1,          // Cyan for descriptions
	Codeblock:      SyntaxString,     // Green for code examples
	Program:        Accent2,          // Magenta for program name
	DimmedArgument: TextMuted,        // Muted for optional args
	Comment:        SyntaxComment,    // Muted for comments
	Flag:           TermGreen,        // Green for flags
	FlagDefault:    TermBrightYellow, // Bright yellow for defaults
	Command:        Accent2,          // Magenta for commands
	QuotedString:   SyntaxString,     // Green for strings
	Argument:       Accent1,          // Cyan for arguments
	Help:           Text,             // Main text for help
	Dash:           TextMuted,        // Muted for separators
	ErrorHeader: [2]color.Color{
		ErrorColor, // Red foreground
		Bg0,        // Dark background
	},
	ErrorDetails: TermBrightRed, // Bright red for error text
}

// getColorScheme is called by Fang to get the color scheme.
func getColorScheme(ld lipgloss.LightDarkFunc) fang.ColorScheme {
	return colorScheme
}
