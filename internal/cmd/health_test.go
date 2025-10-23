package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/sebdah/goldie/v2"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		isError    bool
	}{
		{
			name:       "prints output for no config file",
			configFile: "",
		},
		{
			name:       "prints output for loading config file",
			configFile: "/home/.config/ghost/config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			cmd := &cli.Command{
				Metadata: map[string]any{
					"output":     output,
					"configFile": altsrc.NewStringPtrSourcer(&tt.configFile),
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
