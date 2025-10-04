package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// getUserInput retrieves user input from either piped stdin or command-line arguments.
// It handles both piped input and direct arguments, combining them when both are provided.
// Returns an error if no input is available from either source.
func getUserInput(cmd *cobra.Command, args []string, logger *log.Logger) (string, error) {

	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInput, err)
	}

	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	logger.Debug("detected input mode", "piped", isPiped)

	var query string

	if isPiped {
		query, err = readPipedInput(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrIO, err)
		}

		logger.Debug("read piped input", "bytes", len(query))

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, " ")

			logger.Debug("combined piped input with arguments", "argCount", len(args))
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")

		logger.Debug("using direct arguments as query", "argCount", len(args))
	} else {
		logger.Warn("no input provided")

		return "", fmt.Errorf("%w: provide a query or pipe input", ErrInput)
	}

	return query, nil
}

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
