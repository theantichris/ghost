package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestMatchesCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		binding key.Binding
		want    bool
	}{
		{
			name:    "single key matches",
			cmd:     "q",
			binding: key.NewBinding(key.WithKeys("q")),
			want:    true,
		},
		{
			name:    "single key does not match",
			cmd:     "x",
			binding: key.NewBinding(key.WithKeys("q")),
			want:    false,
		},
		{
			name:    "multi-key matches first",
			cmd:     "j",
			binding: key.NewBinding(key.WithKeys("j", "down")),
			want:    true,
		},
		{
			name:    "multi-key matches second",
			cmd:     "down",
			binding: key.NewBinding(key.WithKeys("j", "down")),
			want:    true,
		},
		{
			name:    "multi-key does not match",
			cmd:     "k",
			binding: key.NewBinding(key.WithKeys("j", "down")),
			want:    false,
		},
		{
			name:    "empty command does not match",
			cmd:     "",
			binding: key.NewBinding(key.WithKeys("q")),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesCommand(tt.cmd, tt.binding); got != tt.want {
				t.Errorf("matchesCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
