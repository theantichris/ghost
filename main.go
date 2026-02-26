package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/v3/cmd"
	"github.com/theantichris/ghost/v3/style"
)

func main() {
	rootCmd, loggerCleanup, err := cmd.NewRootCmd()

	if err != nil {
		style.FangErrorHandler(os.Stderr, fang.Styles{}, err)
		os.Exit(1)
	}

	defer func() {
		_ = loggerCleanup()
	}()

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(rootCmd.Version),
		fang.WithColorSchemeFunc(style.GetFangColorScheme),
		fang.WithErrorHandler(style.FangErrorHandler),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
