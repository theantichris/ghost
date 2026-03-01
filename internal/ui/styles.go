package ui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/v3/style"
)

var textAreaStyles textarea.Styles = textarea.Styles{
	Focused: textarea.StyleState{Base: baseBackground},
	Blurred: textarea.StyleState{Base: baseBackground},
}

var baseBackground lipgloss.Style = lipgloss.NewStyle().Background(style.Bg3)

var viewportStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true).
	BorderForeground(style.Accent0)

var inputStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true).
	BorderForeground(style.Accent0)

var statusBarStyle lipgloss.Style = lipgloss.NewStyle().
	Background(style.Bg3).
	Border(lipgloss.NormalBorder(), true).
	BorderForeground(style.Accent0)
