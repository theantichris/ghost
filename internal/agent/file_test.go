package agent

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}

	// JPEG magic bytes: FF D8 FF
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

	// WebP magic bytes: RIFF + size + WEBP + VP8 chunk
	webpBytes := []byte{
		0x52, 0x49, 0x46, 0x46, // RIFF
		0x24, 0x00, 0x00, 0x00, // file size (non-zero)
		0x57, 0x45, 0x42, 0x50, // WEBP
		0x56, 0x50, 0x38, 0x20, // VP8 (lossy)
	}

	// GIF magic bytes: GIF89a
	gifBytes := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00}

	// Plain text
	textBytes := []byte("Hello, this is plain text content for testing purposes.")

	// JSON content
	jsonBytes := []byte(`{"key": "value", "number": 42}`)

	// XML content
	xmlBytes := []byte(`<?xml version="1.0"?><root><item>test</item></root>`)

	// SVG content (XML with .svg extension)
	svgBytes := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)

	// Binary executable (ELF magic bytes)
	elfBytes := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}

	// Shell script
	shBytes := []byte("#!/bin/bash\necho 'hello world'\n")

	// JavaScript
	jsBytes := []byte("function test() { return 42; }")

	tests := []struct {
		name     string
		filename string
		content  []byte
		isDir    bool
		want     FileType
		wantErr  error
	}{
		{
			name:     "PNG image",
			filename: "test.png",
			content:  pngBytes,
			want:     FileTypeImage,
			wantErr:  nil,
		},
		{
			name:     "JPEG image",
			filename: "test.jpg",
			content:  jpegBytes,
			want:     FileTypeImage,
			wantErr:  nil,
		},
		{
			name:     "WebP image",
			filename: "test.webp",
			content:  webpBytes,
			want:     FileTypeImage,
			wantErr:  nil,
		},
		{
			name:     "GIF image rejected",
			filename: "test.gif",
			content:  gifBytes,
			want:     "",
			wantErr:  ErrFileTypeUnsupported,
		},
		{
			name:     "plain text file",
			filename: "test.txt",
			content:  textBytes,
			want:     FileTypeText,
			wantErr:  nil,
		},
		{
			name:     "JSON file",
			filename: "test.json",
			content:  jsonBytes,
			want:     FileTypeText,
			wantErr:  nil,
		},
		{
			name:     "XML file",
			filename: "test.xml",
			content:  xmlBytes,
			want:     FileTypeText,
			wantErr:  nil,
		},
		{
			name:     "SVG file routes to image",
			filename: "test.svg",
			content:  svgBytes,
			want:     FileTypeImage,
			wantErr:  nil,
		},
		{
			name:     "shell script",
			filename: "test.sh",
			content:  shBytes,
			want:     FileTypeText,
			wantErr:  nil,
		},
		{
			name:     "JavaScript file",
			filename: "test.js",
			content:  jsBytes,
			want:     FileTypeText,
			wantErr:  nil,
		},
		{
			name:     "binary file rejected",
			filename: "test.exe",
			content:  elfBytes,
			want:     "",
			wantErr:  ErrFileTypeUnsupported,
		},
		{
			name:     "directory returns FileTypeDir",
			filename: "testdir",
			isDir:    true,
			want:     FileTypeDir,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)

			if tt.isDir {
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatalf("failed to create test directory: %v", err)
				}
			} else {
				if err := os.WriteFile(path, tt.content, 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			got, err := DetectFileType(path)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("DetectFileType() error = nil, want %v", tt.wantErr)
				}

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DetectFileType() error = %v, want %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("DetectFileType() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("DetectFileType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectFileType_NonExistentFile(t *testing.T) {
	_, err := DetectFileType("/nonexistent/path/to/file.txt")

	if err == nil {
		t.Fatal("DetectFileType() error = nil, want error")
	}

	if !errors.Is(err, ErrFileAccess) {
		t.Errorf("DetectFileType() error = %v, want %v", err, ErrFileAccess)
	}
}

func TestDetectFileType_FileTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large.txt")

	// Create a sparse file larger than 10MB limit.
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Write 1 byte past 10MB to create a file that exceeds the limit.
	if _, err := f.WriteAt([]byte{0x00}, 11*1024*1024); err != nil {
		defer func() { _ = f.Close() }()
		t.Fatalf("failed to write to test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	_, err = DetectFileType(path)

	if err == nil {
		t.Fatal("DetectFileType() error = nil, want error")
	}

	if !errors.Is(err, ErrFileSize) {
		t.Errorf("DetectFileType() error = %v, want %v", err, ErrFileSize)
	}
}

func TestIsImage(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		path      string
		want      bool
	}{
		{"PNG", "image/png", "test.png", true},
		{"JPEG", "image/jpeg", "test.jpg", true},
		{"WebP", "image/webp", "test.webp", true},
		{"SVG with correct extension", "text/xml", "test.svg", true},
		{"XML not SVG", "text/xml", "test.xml", false},
		{"GIF rejected", "image/gif", "test.gif", false},
		{"text not image", "text/plain", "test.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isImage(tt.mediaType, tt.path); got != tt.want {
				t.Errorf("isImage(%q, %q) = %v, want %v", tt.mediaType, tt.path, got, tt.want)
			}
		})
	}
}

func TestIsText(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		want      bool
	}{
		{"text/plain", "text/plain", true},
		{"text/html", "text/html", true},
		{"text/css", "text/css", true},
		{"application/json", "application/json", true},
		{"application/xml", "application/xml", true},
		{"application/javascript", "application/javascript", true},
		{"application/x-sh", "application/x-sh", true},
		{"image/png not text", "image/png", false},
		{"application/octet-stream not text", "application/octet-stream", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isText(tt.mediaType); got != tt.want {
				t.Errorf("isText(%q) = %v, want %v", tt.mediaType, got, tt.want)
			}
		})
	}
}
