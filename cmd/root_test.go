package cmd

import (
	"errors"
	"testing"
)

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
		err     error
	}{
		{
			name:   "does not return error for json",
			format: "json",
		},
		{
			name:   "does not return error for markdown",
			format: "markdown",
		},
		{
			name:   "does not return error for empty format",
			format: "",
		},
		{
			name:    "returns error for invalid format",
			format:  "butts",
			wantErr: true,
			err:     ErrInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormat(tt.format)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateFormat() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("validateFormat() err = %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("validateFormat() error = %v, want no error", err)
			}
		})
	}
}
