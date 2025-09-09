package app

import "testing"

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		_ = New()
	})
}
