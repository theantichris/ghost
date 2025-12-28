package main

import (
	"context"
	"fmt"
	"io"
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

func errorHandler(w io.Writer, styles fang.Styles, err error) {
	errorHeader := styles.ErrorHeader.Render("󱙜")
	errorDetails := styles.ErrorText.Render(err.Error())

	fmt.Fprintf(w, "%s\n%s\n", errorHeader, errorDetails)
}
