package agent

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
)

func newTestLogger() *log.Logger {
	return log.New(os.Stderr)
}

func TestLoadPrompt(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string // returns promptDir
		filename     string
		defaultValue string
		want         string
		wantErr      bool
	}{
		{
			name: "file exists returns content",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("custom prompt"), 0640)
				if err != nil {
					t.Fatalf("setup write: %v", err)
				}
				return dir
			},
			filename:     "test.md",
			defaultValue: "default prompt",
			want:         "custom prompt",
		},
		{
			name: "file does not exist creates file with default",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			filename:     "test.md",
			defaultValue: "default prompt",
			want:         "default prompt",
		},
		{
			name: "write fails on missing file returns ErrPromptLoad",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				err := os.Chmod(dir, 0444)
				if err != nil {
					t.Fatalf("setup chmod: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, 0755) })
				return dir
			},
			filename:     "test.md",
			defaultValue: "default prompt",
			wantErr:      true,
		},
		{
			name: "read error not ErrNotExist returns ErrPromptLoad",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// Create a directory where a file is expected — ReadFile on a dir
				// returns a non-ErrNotExist error.
				err := os.Mkdir(filepath.Join(dir, "test.md"), 0755)
				if err != nil {
					t.Fatalf("setup mkdir: %v", err)
				}
				return dir
			},
			filename:     "test.md",
			defaultValue: "default prompt",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptDir := tt.setup(t)

			got, err := loadPrompt(promptDir, tt.filename, tt.defaultValue)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, ErrPromptLoad) {
					t.Errorf("expected ErrPromptLoad, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadPrompt_FileCreatedOnDisk(t *testing.T) {
	dir := t.TempDir()

	_, err := loadPrompt(dir, "test.md", "default content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}

	if string(data) != "default content" {
		t.Errorf("file content = %q, want %q", string(data), "default content")
	}
}

func TestLoadPrompts(t *testing.T) {
	logger := newTestLogger()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns configDir
		verify  func(t *testing.T, p Prompt, configDir string)
		wantErr bool
	}{
		{
			name: "fresh directory populates defaults and creates files",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			verify: func(t *testing.T, p Prompt, configDir string) {
				t.Helper()
				if p.System != systemPrompt {
					t.Errorf("System = %q, want default", p.System)
				}
				if p.VisionSystem != visionSystemPrompt {
					t.Errorf("VisionSystem = %q, want default", p.VisionSystem)
				}
				if p.Vision != visionPrompt {
					t.Errorf("Vision = %q, want default", p.Vision)
				}
				if p.JSON != jsonPrompt {
					t.Errorf("JSON = %q, want default", p.JSON)
				}
				if p.Markdown != markdownPrompt {
					t.Errorf("Markdown = %q, want default", p.Markdown)
				}

				// Verify files were created on disk.
				promptDir := filepath.Join(configDir, "prompts")
				files := []string{"system.md", "vision_system.md", "vision.md", "json.md", "markdown.md"}
				for _, f := range files {
					if _, err := os.Stat(filepath.Join(promptDir, f)); err != nil {
						t.Errorf("expected file %s to exist: %v", f, err)
					}
				}
			},
		},
		{
			name: "all files exist returns custom content",
			setup: func(t *testing.T) string {
				t.Helper()
				configDir := t.TempDir()
				promptDir := filepath.Join(configDir, "prompts")
				if err := os.MkdirAll(promptDir, 0750); err != nil {
					t.Fatalf("setup mkdir: %v", err)
				}
				files := map[string]string{
					"system.md":        "custom system",
					"vision_system.md": "custom vision system",
					"vision.md":        "custom vision",
					"json.md":          "custom json",
					"markdown.md":      "custom markdown",
				}
				for name, content := range files {
					if err := os.WriteFile(filepath.Join(promptDir, name), []byte(content), 0640); err != nil {
						t.Fatalf("setup write %s: %v", name, err)
					}
				}
				return configDir
			},
			verify: func(t *testing.T, p Prompt, _ string) {
				t.Helper()
				if p.System != "custom system" {
					t.Errorf("System = %q, want %q", p.System, "custom system")
				}
				if p.VisionSystem != "custom vision system" {
					t.Errorf("VisionSystem = %q, want %q", p.VisionSystem, "custom vision system")
				}
				if p.Vision != "custom vision" {
					t.Errorf("Vision = %q, want %q", p.Vision, "custom vision")
				}
				if p.JSON != "custom json" {
					t.Errorf("JSON = %q, want %q", p.JSON, "custom json")
				}
				if p.Markdown != "custom markdown" {
					t.Errorf("Markdown = %q, want %q", p.Markdown, "custom markdown")
				}
			},
		},
		{
			name: "mixed files returns custom for existing and defaults for missing",
			setup: func(t *testing.T) string {
				t.Helper()
				configDir := t.TempDir()
				promptDir := filepath.Join(configDir, "prompts")
				if err := os.MkdirAll(promptDir, 0750); err != nil {
					t.Fatalf("setup mkdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(promptDir, "system.md"), []byte("custom system"), 0640); err != nil {
					t.Fatalf("setup write: %v", err)
				}
				if err := os.WriteFile(filepath.Join(promptDir, "json.md"), []byte("custom json"), 0640); err != nil {
					t.Fatalf("setup write: %v", err)
				}
				return configDir
			},
			verify: func(t *testing.T, p Prompt, _ string) {
				t.Helper()
				if p.System != "custom system" {
					t.Errorf("System = %q, want %q", p.System, "custom system")
				}
				if p.VisionSystem != visionSystemPrompt {
					t.Errorf("VisionSystem = %q, want default", p.VisionSystem)
				}
				if p.Vision != visionPrompt {
					t.Errorf("Vision = %q, want default", p.Vision)
				}
				if p.JSON != "custom json" {
					t.Errorf("JSON = %q, want %q", p.JSON, "custom json")
				}
				if p.Markdown != markdownPrompt {
					t.Errorf("Markdown = %q, want default", p.Markdown)
				}
			},
		},
		{
			name: "directory creation fails returns ErrPromptLoad",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				readOnly := filepath.Join(dir, "readonly")
				if err := os.Mkdir(readOnly, 0444); err != nil {
					t.Fatalf("setup mkdir: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(readOnly, 0755) })
				// MkdirAll will fail trying to create "prompts" under a read-only dir.
				return filepath.Join(readOnly, "nested")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDir := tt.setup(t)

			got, err := LoadPrompts(configDir, logger)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, ErrPromptLoad) {
					t.Errorf("expected ErrPromptLoad, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tt.verify(t, got, configDir)
		})
	}
}
