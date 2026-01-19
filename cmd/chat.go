package cmd

import "github.com/spf13/cobra"

const (
	chatUseText     = "chat"
	chatShortText   = "starts ghost in chat mode"
	chatLongText    = "starts ghost in chat mode, use :q to quit"
	chatExampleText = "ghost chat"
)

func newChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     chatUseText,
		Short:   chatShortText,
		Long:    chatLongText,
		Example: chatExampleText,
		Args:    cobra.NoArgs,
		RunE:    runChat,
	}

	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	return nil
}
