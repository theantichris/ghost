package tui

import (
	"testing"
	"time"

	"github.com/theantichris/ghost/v3/internal/storage"
)

func TestThreadItem_Title(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "returns thread title",
			title: "tell me a joke",
			want:  "tell me a joke",
		},
		{
			name:  "returns empty title",
			title: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := threadItem{
				thread: storage.Thread{Title: tt.title},
			}

			if got := item.Title(); got != tt.want {
				t.Errorf("Title() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThreadItem_Description(t *testing.T) {
	tests := []struct {
		name      string
		updatedAt time.Time
		want      string
	}{
		{
			name:      "formats timestamp with ANSIC layout",
			updatedAt: time.Date(2026, 2, 15, 18, 59, 18, 0, time.UTC),
			want:      time.Date(2026, 2, 15, 18, 59, 18, 0, time.UTC).Format(time.ANSIC),
		},
		{
			name:      "formats zero time",
			updatedAt: time.Time{},
			want:      time.Time{}.Format(time.ANSIC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := threadItem{
				thread: storage.Thread{UpdatedAt: tt.updatedAt},
			}

			if got := item.Description(); got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThreadItem_FilterValue(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "returns thread title for filtering",
			title: "tell me a joke",
			want:  "tell me a joke",
		},
		{
			name:  "returns empty string for empty title",
			title: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := threadItem{
				thread: storage.Thread{Title: tt.title},
			}

			if got := item.FilterValue(); got != tt.want {
				t.Errorf("FilterValue() = %q, want %q", got, tt.want)
			}
		})
	}
}
