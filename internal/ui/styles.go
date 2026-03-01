package ui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/v3/style"
)

var (
	baseBackground lipgloss.Style  = lipgloss.NewStyle().Background(style.Bg3)
	textAreaStyles textarea.Styles = textarea.Styles{
		Focused: textarea.StyleState{Base: baseBackground},
		Blurred: textarea.StyleState{Base: baseBackground},
	}

	panelStyle lipgloss.Style = lipgloss.NewStyle().
			Background(style.Bg3).
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(style.Accent0)
)
