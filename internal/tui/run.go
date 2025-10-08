package tui

import tea "github.com/charmbracelet/bubbletea"

func Run(model Model) error {
	_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()

	return err
}
