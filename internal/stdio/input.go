package stdio

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// InputReader handles user input retrieval with injectable stdin detection.
type InputReader struct {
	logger        *log.Logger
	stdinDetector func() (bool, error)
}

// NewInputReader creates an InputReader with the default stdin detector.
func NewInputReader(logger *log.Logger) *InputReader {
	inputReader := InputReader{
		logger:        logger,
		stdinDetector: isPiped,
	}

	return &inputReader
}

func isPiped() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	return (stat.Mode() & os.ModeCharDevice) == 0, nil
}

// Read retrieves user input from either piped stdin or command-line arguments,
// combining them when both are provided. Read returns an error if no input is
// available from either source.
func (inputReader *InputReader) Read(cmd *cobra.Command, args []string) (string, error) {
	isPiped, err := inputReader.stdinDetector()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIO, err)
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

		return "", fmt.Errorf("%w: provide a query or pipe input", ErrIO)
	}

	return query, nil
}

// readPipedInput reads all input from the provided reader until EOF.
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
