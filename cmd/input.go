package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// readPipedInput reads all input from the provided reader until EOF.
// It's used to capture piped input from stdin.
func readPipedInput(input io.Reader) (string, error) {
	reader := bufio.NewReader(input)

	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					lines = append(lines, line)
				}

				break
			}

			return "", fmt.Errorf("%w: %w", ErrIO, err)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, ""), nil
}
