package cmd

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("prints ghost system online", func(t *testing.T) {
		t.Parallel()

		err := Run(context.Background(), []string{})
		if err != nil {
			t.Fatalf("expect no error got, %v", err)
		}
	})
}
