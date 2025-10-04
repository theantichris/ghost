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

// inputReader handles user input retrieval with injectable stdin detection.
type inputReader struct {
	logger        *log.Logger
	stdinDetector func() (bool, error)
}

// newInputReader creates an inputReader with the default stdin detector.
func newInputReader(logger *log.Logger) *inputReader {
	inputReader := inputReader{
		logger: logger,
		stdinDetector: func() (bool, error) {
			stat, err := os.Stdin.Stat()
			if err != nil {
				return false, err
			}

			return (stat.Mode() & os.ModeCharDevice) == 0, nil
		},
	}

	return &inputReader
}

// read retrieves user input from either piped stdin or command-line arguments.
// It handles both piped input and direct arguments, combining them when both are provided.
// Returns an error if no input is available from either source.
func (inputReader *inputReader) read(cmd *cobra.Command, args []string) (string, error) {
	isPiped, err := inputReader.stdinDetector()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInput, err)
	}

	inputReader.logger.Debug("detected input mode", "piped", isPiped)

	var query string

	if isPiped {
		query, err = readPipedInput(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrIO, err)
		}

		inputReader.logger.Debug("read piped input", "bytes", len(query))

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, " ")

			inputReader.logger.Debug("combined piped input with arguments", "argCount", len(args))
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")

		inputReader.logger.Debug("using direct arguments as query", "argCount", len(args))
	} else {
		inputReader.logger.Warn("no input provided")

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
