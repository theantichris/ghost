package ui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/v3/style"
)

// horizontalChrome is the size of horizontal borders, margin, and padding.
const horizontalChrome = 2

// verticalChrome is the size of vertical borders, margin, and padding.
const verticalChrome = 6

var baseBackground lipgloss.Style = lipgloss.NewStyle().Background(style.Bg3)

var viewportStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true)

var textAreaStyles textarea.Styles = textarea.Styles{
	Focused: textarea.StyleState{Base: baseBackground},
	Blurred: textarea.StyleState{Base: baseBackground},
}

var inputStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true)

var statusBarStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true)

	// height - input height - status bar height = viewport height
