package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/urfave/cli/v3"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name    string
		isError bool
	}{
		{name: "prints output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			cmd := &cli.Command{
				Metadata: map[string]any{
					"output": output,
				},
			}

			err := health(context.Background(), cmd)
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			g := goldie.New(t)
			g.Assert(t, t.Name(), output.Bytes())
		})
	}
}
