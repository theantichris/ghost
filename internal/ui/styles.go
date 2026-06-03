package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/theantichris/ghost/v3/style"
)

var (
	panelStyle lipgloss.Style = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder(), true).
		BorderForeground(style.Accent0)
)
