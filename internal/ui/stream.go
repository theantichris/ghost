package ui

import tea "charm.land/bubbletea/v2"

// LLMDoneMsg signals the LLM request is complete.
type LLMDoneMsg struct{}

// listenForChunk returns a command that waits for the next chunk from the channel.
func listenForChunk(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return LLMDoneMsg{}
		}

		return msg
	}
}
