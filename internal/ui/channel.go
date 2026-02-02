package ui

import tea "charm.land/bubbletea/v2"

// ListenForChunk returns a command that waits for the next chunk from the channel.
func ListenForChunk(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return LLMDoneMsg{}
		}

		return msg
	}
}
