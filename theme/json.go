package theme

import (
	"strings"
	"unicode"
)

// JSON highlights JSON text with colors using lipgloss.
func JSON(input string) string {
	var result strings.Builder
	result.Grow(len(input) * 2)

	inString := false
	inKey := false
	escaped := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// Handle escape sequences
		if escaped {
			result.WriteByte(ch)
			escaped = false
			continue
		}

		if inString && ch == '\\' {
			result.WriteByte(ch)
			escaped = true
			continue
		}

		// Handle strings
		if ch == '"' {
			if !inString {
				// Starting a string - determine if it's a key
				inString = true
				// Look ahead to see if this is followed by a colon (key)
				j := i + 1
				for j < len(input) && input[j] != '"' {
					if input[j] == '\\' {
						j++ // skip escaped char
					}
					j++
				}
				if j < len(input) {
					// Found closing quote, check for colon
					k := j + 1
					for k < len(input) && unicode.IsSpace(rune(input[k])) {
						k++
					}
					inKey = k < len(input) && input[k] == ':'
				}

				// Capture the string content
				stringStart := i
				i++
				for i < len(input) && input[i] != '"' {
					if input[i] == '\\' && i+1 < len(input) {
						i++ // skip escaped char
					}
					i++
				}

				if i < len(input) {
					stringContent := input[stringStart : i+1]
					if inKey {
						result.WriteString(JSONKey.Render(stringContent))
					} else {
						result.WriteString(JSONString.Render(stringContent))
					}
					inString = false
					inKey = false
				}
				continue
			}
		}

		if inString {
			result.WriteByte(ch)
			continue
		}

		// Handle numbers
		if unicode.IsDigit(rune(ch)) || (ch == '-' && i+1 < len(input) && unicode.IsDigit(rune(input[i+1]))) {
			numStart := i
			i++
			for i < len(input) {
				ch := input[i]
				if unicode.IsDigit(rune(ch)) || ch == '.' || ch == 'e' || ch == 'E' || ch == '+' || ch == '-' {
					i++
				} else {
					break
				}
			}
			result.WriteString(JSONNumber.Render(input[numStart:i]))
			i--
			continue
		}

		// Handle booleans and null
		if ch == 't' && i+3 < len(input) && input[i:i+4] == "true" {
			result.WriteString(JSONBoolNull.Render("true"))
			i += 3
			continue
		}
		if ch == 'f' && i+4 < len(input) && input[i:i+5] == "false" {
			result.WriteString(JSONBoolNull.Render("false"))
			i += 4
			continue
		}
		if ch == 'n' && i+3 < len(input) && input[i:i+4] == "null" {
			result.WriteString(JSONBoolNull.Render("null"))
			i += 3
			continue
		}

		// Handle structural characters
		switch ch {
		case '{', '}', '[', ']':
			result.WriteString(JSONBracket.Render(string(ch)))
		case ':', ',':
			result.WriteString(JSONPunctuation.Render(string(ch)))
		default:
			result.WriteByte(ch)
		}
	}

	return result.String()
}
