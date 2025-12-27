package main

import (
	"context"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

func main() {
	if err := fang.Execute(
		context.Background(),
		cmd.RootCmd,
		fang.WithVersion(cmd.Version),
		fang.WithColorSchemeFunc(cyberpunkTheme),
	); err != nil {
		os.Exit(1)
	}
}

func cyberpunkTheme(ld lipgloss.LightDarkFunc) fang.ColorScheme {
	theme := fang.ColorScheme{
		Title:       lipgloss.Color("#FF00FF"),
		Description: lipgloss.Color("#00FFFF"),
		Flag:        lipgloss.Color("#00FF00"),
		Command:     lipgloss.Color("#FF0080"),
		Argument:    lipgloss.Color("#80FF00"),
	}

	return theme
}
