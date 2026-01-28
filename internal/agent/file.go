package agent

import (
	"errors"
	"fmt"
	"os"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

var (
	ErrFileAccess = errors.New("failed to access file")
	ErrIsDir      = errors.New("path is a directory, not a file")
	ErrFileSize   = errors.New("file exceeds 10MB limit")
	ErrReadFile   = errors.New("failed to read file")
)

// ReadFileForContext reads a file and returns formatted content for the LLM.
// Returns the formatted content and any error encountered.
func ReadFileForContext(path string) (string, error) {
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
		return "", fmt.Errorf("%w, %w", ErrReadFile, err)
	}

	return fmt.Sprintf("[FILE: %s]\n%s", path, string(content)), nil
}
