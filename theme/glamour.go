package theme

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/glamour/ansi"
)

// CyberpunkTheme returns a custom Glamour theme based on the Ghost cyberpunk aesthetic.
func CyberpunkTheme() ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       colorPtr(Text),
			},
			Margin: uintPtr(2),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: colorPtr(TextMuted),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       colorPtr(Accent1),
				Bold:        boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           colorPtr(Accent1),
				BackgroundColor: colorPtr(Bg0),
				Bold:            boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "## ",
				Color:  colorPtr(Accent1),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "### ",
				Color:  colorPtr(Accent1),
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "#### ",
				Color:  colorPtr(Accent1),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "##### ",
				Color:  colorPtr(Accent1),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "###### ",
				Color:  colorPtr(Accent1),
				Bold:   boolPtr(false),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Italic: boolPtr(true),
			Color:  colorPtr(Accent2),
		},
		Strong: ansi.StylePrimitive{
			Bold:  boolPtr(true),
			Color: colorPtr(Text),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  stringPtr("#029fec91"),
			Format: "\n--------\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: ansi.StyleTask{
			Ticked:   "[✓] ",
			Unticked: "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     colorPtr(Link),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: colorPtr(Accent1),
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     colorPtr(Accent2),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  colorPtr(TextMuted),
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           colorPtr(Accent0),
				BackgroundColor: stringPtr("#001a24"),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: colorPtr(Accent0),
				},
				Margin: uintPtr(2),
			},
			Chroma: &ansi.Chroma{
				Text:                ansi.StylePrimitive{Color: colorPtr(Text)},
				Error:               ansi.StylePrimitive{Color: colorPtr(Error), BackgroundColor: colorPtr(Bg0)},
				Comment:             ansi.StylePrimitive{Color: colorPtr(SyntaxComment)},
				CommentPreproc:      ansi.StylePrimitive{Color: colorPtr(Accent2)},
				Keyword:             ansi.StylePrimitive{Color: colorPtr(SyntaxKeyword)},
				KeywordReserved:     ansi.StylePrimitive{Color: colorPtr(Accent2)},
				KeywordNamespace:    ansi.StylePrimitive{Color: colorPtr(Accent3)},
				KeywordType:         ansi.StylePrimitive{Color: colorPtr(SyntaxType)},
				Operator:            ansi.StylePrimitive{Color: colorPtr(SyntaxOperator)},
				Punctuation:         ansi.StylePrimitive{Color: colorPtr(Text)},
				Name:                ansi.StylePrimitive{Color: colorPtr(SyntaxFunction)},
				NameBuiltin:         ansi.StylePrimitive{Color: colorPtr(SyntaxFunction)},
				NameTag:             ansi.StylePrimitive{Color: colorPtr(Accent1)},
				NameAttribute:       ansi.StylePrimitive{Color: colorPtr(Accent0)},
				NameClass:           ansi.StylePrimitive{Color: colorPtr(SyntaxType)},
				NameConstant:        ansi.StylePrimitive{Color: colorPtr(Accent0)},
				NameDecorator:       ansi.StylePrimitive{Color: colorPtr(Accent2)},
				NameException:       ansi.StylePrimitive{Color: colorPtr(Error)},
				NameFunction:        ansi.StylePrimitive{Color: colorPtr(SyntaxFunction)},
				NameOther:           ansi.StylePrimitive{Color: colorPtr(SyntaxFunction)},
				Literal:             ansi.StylePrimitive{Color: colorPtr(SyntaxNumber)},
				LiteralNumber:       ansi.StylePrimitive{Color: colorPtr(SyntaxNumber)},
				LiteralDate:         ansi.StylePrimitive{Color: colorPtr(SyntaxNumber)},
				LiteralString:       ansi.StylePrimitive{Color: colorPtr(SyntaxString)},
				LiteralStringEscape: ansi.StylePrimitive{Color: colorPtr(Accent2)},
				GenericDeleted:      ansi.StylePrimitive{Color: colorPtr(Error)},
				GenericEmph:         ansi.StylePrimitive{Italic: boolPtr(true)},
				GenericInserted:     ansi.StylePrimitive{Color: colorPtr(Success)},
				GenericStrong:       ansi.StylePrimitive{Bold: boolPtr(true)},
				GenericSubheading:   ansi.StylePrimitive{Color: colorPtr(Accent2), Bold: boolPtr(true)},
				Background:          ansi.StylePrimitive{BackgroundColor: colorPtr(Bg0)},
			},
		},
		Table: ansi.StyleTable{
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionTerm: ansi.StylePrimitive{
			Color: colorPtr(Accent2),
			Bold:  boolPtr(true),
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n🠶 ",
			Color:       colorPtr(TextMuted),
		},
	}
}

// Helper functions to create pointers.
func colorPtr(c color.Color) *string {
	// lipgloss.Color returns a color.Color that can be a hex string
	// We need to extract the underlying string value
	// Since lipgloss stores colors as hex strings internally, we can format it
	if adaptiveColor, ok := c.(interface{ ANSI256() string }); ok {
		s := adaptiveColor.ANSI256()
		return &s
	}
	// Fallback: convert to hex string
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	s := fmt.Sprintf("#%02X%02X%02X", r8, g8, b8)
	return &s
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func uintPtr(u uint) *uint {
	return &u
}
