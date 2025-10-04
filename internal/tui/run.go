package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	_, err := tea.NewProgram(Model{}, tea.WithAltScreen()).Run()

	return err
}
