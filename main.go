package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

func main() {
	// TODO: Can I use this context in the commands?
	if err := fang.Execute(context.Background(), cmd.Execute()); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)

		os.Exit(1)
	}
}
