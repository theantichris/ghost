package main

import (
	"errors"
	"testing"
)

func TestGetPrompt(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
		err      error
	}{
		{
			name:     "returns prompt",
			args:     []string{"ghost", "tell me a joke"},
			expected: "tell me a joke",
		},
		{
			name:    "returns error for no prompt",
			args:    []string{"ghost"},
			wantErr: true,
			err:     errPromptNotDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getPrompt(tt.args)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if actual != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, actual)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
			}
		})
	}
}
