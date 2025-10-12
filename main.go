// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

// main initializes and executes the Ghost CLI application using the fang framework.
func main() {
	cmd := &cli.Command{
		Name:  "ghost",
		Usage: "send ghost a prompt",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("ghost system online")

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
