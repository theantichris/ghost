package agent

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type FileType string

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB

	FileTypeText  FileType = "text"
	FileTypeImage FileType = "image"
	FileTypeDir   FileType = "dir"
)

var (
	ErrFileAccess          = errors.New("failed to access file")
	ErrIsDir               = errors.New("path is a directory, not a file")
	ErrFileSize            = errors.New("file exceeds 10MB limit")
	ErrReadFile            = errors.New("failed to read file")
	ErrFileTypeUnsupported = errors.New("file type unsupported")

	textFileTypes = []string{
		"application/javascript",
		"application/json",
		"application/x-sh",
		"application/xml",
	}

	imageFileTypes = []string{
		"image/png",
		"image/jpeg",
		"image/webp",
	}
)

// DetectFileType returns a FileType based on files mime type and path.
// Returns FileTypeDir if path is a directory.
// Returns ErrFileTypeUnSupported if the file type is not supported.
func DetectFileType(path string) (FileType, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFileAccess, err)
	}

	if info.IsDir() {
		return FileTypeDir, nil
	}

	if info.Size() > maxFileSize {
		return "", fmt.Errorf("%w (%d bytes)", ErrFileSize, info.Size())
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadFile, err)
	}
	defer func() { _ = file.Close() }()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadFile, err)
	}

	mime := http.DetectContentType(buffer[:n])
	mediaType := strings.SplitN(mime, ";", 2)[0]

	if isImage(mediaType, path) {
		return FileTypeImage, nil
	}

	if isText(mediaType) {
		return FileTypeText, nil
	}

	return "", ErrFileTypeUnsupported
}

func isImage(mediaType, path string) bool {
	if slices.Contains(imageFileTypes, mediaType) {
		return true
	}

	if mediaType == "text/xml" && filepath.Ext(path) == ".svg" {
		return true
	}

	return false
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
		return "", fmt.Errorf("%w: %w", ErrReadFile, err)
	}

	return fmt.Sprintf("[FILE: %s]\n%s", path, string(content)), nil
}
