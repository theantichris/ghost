package agent

import "testing"

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
