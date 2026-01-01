package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
	"github.com/theantichris/ghost/theme"
)

func main() {
	rootCmd, loggerCleanup, err := cmd.NewRootCmd()

	if err != nil {
		theme.FangErrorHandler(os.Stderr, fang.Styles{}, err)
		os.Exit(1)
	}

	defer func() {
		_ = loggerCleanup()
	}()

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(rootCmd.Version),
		fang.WithColorSchemeFunc(theme.GetFangColorScheme),
		fang.WithErrorHandler(theme.FangErrorHandler),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
