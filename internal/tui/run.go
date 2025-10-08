package tui

import tea "github.com/charmbracelet/bubbletea"

// Run starts the BubbleTea TUI program with the provided model.
// It runs in alternate screen mode and returns any error that occurs.
func Run(model Model) error {
	_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()

	return err
}
