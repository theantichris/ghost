package ui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/v3/style"
)

var (
	baseBackground lipgloss.Style  = lipgloss.NewStyle().Background(style.Bg2)
	textAreaStyles textarea.Styles = textarea.Styles{
		Focused: textarea.StyleState{Base: baseBackground},
		Blurred: textarea.StyleState{Base: baseBackground},
	}

	panelStyle lipgloss.Style = lipgloss.NewStyle().
			Margin(0, 1).
			Border(lipgloss.NormalBorder(), true).
			Background(style.Bg2).
			BorderForeground(style.Accent0)
)
