package tui

import tea "github.com/charmbracelet/bubbletea"

// streamingChunkMsg carries a single token from the LLM stream.
type streamingChunkMsg struct {
	content string
	sub     <-chan tea.Msg
}

// streamCompleteMsg signals that streaming is complete and carries the full accumulated response.
type streamCompleteMsg struct {
	content string
}

// streamErrorMsg carries error information when an LLM request fails.
type streamErrorMsg struct {
	err error
}
