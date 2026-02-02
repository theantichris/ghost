package agent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
)

var ErrPipedInput = errors.New("data stream interrupted")

// ReadTextFile reads a file and returns formatted content for the LLM.
// Returns the formatted content and any error encountered.
func ReadTextFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFileAccess, err)
	}

	if info.IsDir() {
		return "", ErrIsDir
	}

	if info.Size() > maxFileSize {
		return "", fmt.Errorf("%w (%d bytes)", ErrFileSize, info.Size())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadFile, err)
	}

	return fmt.Sprintf("[FILE: %s]\n%s", path, string(content)), nil
}

func isText(mediaType string) bool {
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	if slices.Contains(textFileTypes, mediaType) {
		return true
	}

	return false
}

// GetPipedInput detects, reads, and returns any input piped to the command.
func GetPipedInput(file *os.File, logger *log.Logger) (string, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return "", nil
	}

	if fileInfo.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	pipedInput, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrPipedInput, err)
	}

	input := strings.TrimSpace(string(pipedInput))

	if len(input) > 0 {
		logger.Debug("intercepted data stream", "size_bytes", len(input))
	}

	return input, nil
}
