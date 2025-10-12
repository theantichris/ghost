package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

type command struct {
	Name  string
	Usage string
}

type commandList map[string]command

var commands = commandList{
	"ghost": {Name: "ghost", Usage: "send ghost a prompt"},
}

func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:  commands["ghost"].Name,
		Usage: commands["ghost"].Usage,
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("ghost system online")

			return nil
		},
	}

	return cmd.Run(ctx, args)
}
