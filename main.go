package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

func main() {
	// TODO: Can I use this context in the commands?
	if err := fang.Execute(context.Background(), cmd.Execute()); err != nil {
		os.Exit(1)
	}
}
