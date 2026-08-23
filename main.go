package main

import (
	"fmt"
	"io"
	"os"
)

const bootstrapMsg = "ghost // bootstrap online"

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ghost // bootstrap failure: %v\n", err)
		os.Exit(1)
	}
}

func run(output io.Writer) error {
	if _, err := fmt.Fprintln(output, bootstrapMsg); err != nil {
		return fmt.Errorf("write bootstrap message: %w", err)
	}

	return nil
}
