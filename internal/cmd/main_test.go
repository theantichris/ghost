package cmd

import (
	"context"
	"io"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/urfave/cli/v3"
)

func TestBefore(t *testing.T) {
	logger := log.New(io.Discard)

	cmd := cli.Command{}
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "host",
			Value: "http://test.dev",
		},
		&cli.StringFlag{
			Name:  "model",
			Value: "default:model",
		},
		&cli.StringFlag{
			Name:  "vision-model",
			Value: "vision:model",
		},
	}
	cmd.Metadata = map[string]any{
		"logger": logger,
	}

	_, err := before(context.Background(), &cmd)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
