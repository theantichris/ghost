package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

func main() {
	if err := fang.Execute(
		context.Background(),
		cmd.RootCmd,
		fang.WithVersion(cmd.Version),
		fang.WithColorSchemeFunc(getColorScheme),
		fang.WithErrorHandler(errorHandler),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
