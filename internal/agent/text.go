package agent

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

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
