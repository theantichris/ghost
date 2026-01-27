package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
)

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

// startLLMStream starts the LLM call in a go routine.
// It returns the first listenForChunk command to start receiving.
func (model *ChatModel) startLLMStream() tea.Cmd {
	model.logger.Debug("transmitting to neural network", "model", model.model, "messages", len(model.messages))

	model.responseCh = make(chan tea.Msg)

	go func() {
		ch := model.responseCh

		messages, err := agent.RunToolLoop(model.ctx, model.toolRegistry, model.url, model.model, model.messages, model.logger)
		if err != nil {
			ch <- LLMErrorMsg{Err: err}
			close(ch)

			return
		}

		model.messages = messages

		_, err = llm.StreamChat(
			model.ctx,
			model.url,
			model.model,
			model.messages,
			nil,
			func(chunk string) {
				ch <- LLMResponseMsg(chunk)
			},
		)

		if err != nil {
			ch <- LLMErrorMsg{Err: err}
		}

		close(ch)
	}()

	return listenForChunk(model.responseCh)
}

func (model ChatModel) handleLLMResponseMsg(msg LLMResponseMsg) (tea.Model, tea.Cmd) {
	model.chatHistory += string(msg)
	model.currentResponse += string(msg)
	model.viewport.SetContent(model.renderHistory())
	model.viewport.GotoBottom()

	return model, listenForChunk(model.responseCh)
}

func (model ChatModel) handleLLMDoneMsg() (tea.Model, tea.Cmd) {
	model.logger.Debug("transmission complete", "response_length", len(model.currentResponse))

	model.chatHistory += "\n\n"
	model.viewport.SetContent(model.renderHistory())
	model.messages = append(model.messages, llm.ChatMessage{
		Role:    llm.RoleAssistant,
		Content: model.currentResponse,
	})

	model.currentResponse = ""

	return model, nil
}

func (model ChatModel) handleLLMErrorMsg(msg LLMErrorMsg) (tea.Model, tea.Cmd) {
	model.logger.Error("neural link disrupted", "error", msg.Err)

	model.chatHistory += fmt.Sprintf("\n[󱙝 error: %v]\n", msg.Err)
	model.viewport.SetContent(model.renderHistory())

	return model, nil
}
