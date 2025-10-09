// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/theantichris/ghost/cmd"
)

// main initializes and executes the Ghost CLI application using the fang framework.
func main() {
	if err := fang.Execute(context.Background(), cmd.Execute()); err != nil {
		os.Exit(1)
	}
}
