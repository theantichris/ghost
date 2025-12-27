package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

func main() {
	if err := fang.Execute(context.Background(), cmd.RootCmd, fang.WithVersion(cmd.Version)); err != nil {
		os.Exit(1)
	}
}
