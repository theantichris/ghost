package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	if _, err := exec.LookPath("ollama"); err != nil {
		t.Skip("ollama not found in PATH, skipping E2E tests")
	}

	ghostBinary, cleanup := buildGhostBinary(t)
	defer cleanup()

	tests := []struct {
		name     string
		stdin    string
		args     []string
		validate func(t *testing.T, stdout, stderr string, exitCode int)
	}{
		{
			name: "basic command and prompt",
			args: []string{"Say the word 'hello'"},
			validate: func(t *testing.T, stdout, stderr string, exitCode int) {
				if exitCode != 0 {
					t.Errorf("expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
				}

				if strings.TrimSpace(stdout) == "" {
					t.Errorf("expected response, got empty string")
				}

				lowerOut := strings.ToLower(stdout)
				hasGreeting := strings.Contains(lowerOut, "hello")

				if !hasGreeting {
					t.Errorf("response didn't contain hello: %s", stdout)
				}
			},
		},
		{
			name:  "piped input",
			args:  []string{"What city is mentioned?"},
			stdin: "Nashville is the capital of Tennessee.",
			validate: func(t *testing.T, stdout, stderr string, exitCode int) {
				if exitCode != 0 {
					t.Errorf("expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
				}

				if strings.TrimSpace(stdout) == "" {
					t.Errorf("expected response, got empty string")
				}

				lowerOut := strings.ToLower(stdout)

				if !strings.Contains(lowerOut, "nashville") {
					t.Errorf("response did not mention 'Nashville': %s", stdout)
				}
			},
		},
		{
			name: "images",
			args: []string{"describe this image in one word."},
			validate: func(t *testing.T, stdout, stderr string, exitCode int) {
				if exitCode != 0 {
					t.Errorf("expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
				}

				if strings.TrimSpace(stdout) == "" {
					t.Errorf("expected response, got empty string")
				}
			},
		},
		{
			name: "health command",
			args: []string{"health"},
			validate: func(t *testing.T, stdout, stderr string, exitCode int) {
				if exitCode != 0 {
					t.Errorf("expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
				}

				if strings.TrimSpace(stdout) == "" {
					t.Errorf("expected response, got empty string")
				}

				lowerOut := strings.ToLower(stdout)

				hasSystemConfig := strings.Contains(lowerOut, "system config") ||
					strings.Contains(lowerOut, "host") ||
					strings.Contains(lowerOut, "model")

				if !hasSystemConfig {
					t.Errorf("output missing system config: %s", stdout)
				}

				hasStatus := strings.Contains(lowerOut, "status") ||
					strings.Contains(lowerOut, "connected") ||
					strings.Contains(lowerOut, "online")

				if !hasStatus {
					t.Errorf("output missing status: %s", stdout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(ghostBinary, tt.args...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			exitCode := 0

			if tt.stdin != "" {
				stdinPipe, err := cmd.StdinPipe()
				if err != nil {
					t.Fatalf("failed to create stdin pipe: %v", err)
				}

				if err := cmd.Start(); err != nil {
					t.Fatalf("failed to start command: %v", err)
				}

				if _, err := stdinPipe.Write([]byte(tt.stdin)); err != nil {
					t.Fatalf("failed to write to stdin: %v", err)
				}

				if err := stdinPipe.Close(); err != nil {
					t.Fatalf("failed to close stdin: %v", err)
				}

				err = cmd.Wait()
				if err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						exitCode = exitErr.ExitCode()
					}
				}
			} else {
				err := cmd.Run()
				if err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						exitCode = exitErr.ExitCode()
					}
				}
			}

			tt.validate(t, stdout.String(), stderr.String(), exitCode)
		})
	}
}

func buildGhostBinary(t *testing.T) (string, func()) {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), "ghost")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)

	var stderr bytes.Buffer
	buildCmd.Stderr = &stderr

	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build ghost: %vnStderr: %s", err, stderr.String())
	}

	cleanup := func() {
		_ = os.Remove(binaryPath)
	}

	return binaryPath, cleanup
}
