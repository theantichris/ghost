package agent

import (
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/log"
)

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

func TestGetPipedInput(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "trims whitespace",
			content: "  hello world  \n",
			want:    "hello world",
		},
		{
			name:    "empty input",
			content: "",
			want:    "",
		},
		{
			name:    "multiline input",
			content: "line1\nline2\nline3",
			want:    "line1\nline2\nline3",
		},
		{
			name:    "only whitespace",
			content: "   \n\t\n  ",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			tmpFile, err := os.CreateTemp("", "ghost-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			defer func(tempFile *os.File) {
				_ = os.Remove(tempFile.Name())
			}(tmpFile)

			defer func(tempFile *os.File) {
				_ = tmpFile.Close()
			}(tmpFile)

			_, err = tmpFile.WriteString(tt.content)
			if err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			_, err = tmpFile.Seek(0, 0)
			if err != nil {
				t.Fatalf("Failed to seek temp file: %v", err)
			}

			got, err := GetPipedInput(tmpFile, logger)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetPipedInput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("GetPipedInput() = %q, want %q", got, tt.want)
			}
		})
	}
}
