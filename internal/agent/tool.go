package agent

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/tool"
)

// RunToolLoop sends a request to the LLM and executes any tool calls needed.
// Results are appended to messages and returned.
// Returns early if no tools are registered.
func RunToolLoop(ctx context.Context, registry tool.Registry, url, model string, messages []llm.ChatMessage, logger *log.Logger) ([]llm.ChatMessage, error) {
	tools := registry.Definitions()
	if len(tools) == 0 {
		logger.Debug("no tools registered, exiting tool loop")

		return messages, nil
	}

	for {
		resp, err := llm.Chat(ctx, url, model, messages, tools)
		if err != nil {
			logger.Error("tool request failed", "error", err)

			return messages, err
		}

		if len(resp.ToolCalls) == 0 {
			// Final response not appended, caller streams it separately for UX
			break
		}

		messages = append(messages, resp)

		for _, toolCall := range resp.ToolCalls {
			logger.Debug("executing tool", "name", toolCall.Function.Name)

			result, err := registry.Execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				logger.Error("tool execution failed", "name", toolCall.Function.Name, "error", err)
				result = fmt.Sprintf("error: %s", err.Error())
			}

			messages = append(messages, llm.ChatMessage{Role: llm.RoleTool, Content: result})
		}
	}

	return messages, nil
}
