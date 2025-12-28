package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
	"github.com/theantichris/ghost/theme"
)

func main() {
	if err := fang.Execute(
		context.Background(),
		cmd.RootCmd,
		fang.WithVersion(cmd.Version),
		fang.WithColorSchemeFunc(theme.GetFangColorScheme),
		fang.WithErrorHandler(theme.FangErrorHandler),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
