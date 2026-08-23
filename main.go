package main

import (
	"fmt"
	"io"
	"os"
)

const bootstrapMsg = "ghost // bootstrap online"

func main() {
	err := run(os.Stdout)

	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(output io.Writer) error {
	if _, err := fmt.Fprintln(output, bootstrapMsg); err != nil {
		return fmt.Errorf("write bootstrap message: %w", err)
	}

	return nil
}
