package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/theantichris/ghost/v4/internal/conversation"
	"github.com/theantichris/ghost/v4/internal/ollama"
)

var (
	errModelRequired  = errors.New("model is required: use --model <name>")
	errPromptRequired = errors.New("prompt is required")
)

type chatFunc func(context.Context, string, []conversation.Message) (conversation.Message, error)

func main() {
	client, err := ollama.NewClient(ollama.DefaultURL, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ghost // uplink initialization failed: %v\n", err)
		os.Exit(1)
	}

	if err := run(context.Background(), os.Args[1:], os.Stdout, client.Chat); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ghost // uplink failure: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer, chat chatFunc) error {
	flags := flag.NewFlagSet("ghost", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	model := flags.String("model", "", "Ollama model")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse command arguments: %w", err)
	}

	modelName := strings.TrimSpace(*model)
	if modelName == "" {
		return errModelRequired
	}

	prompt := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if prompt == "" {
		return errPromptRequired
	}

	response, err := chat(ctx, modelName, []conversation.Message{{Role: conversation.RoleUser, Content: prompt}})
	if err != nil {
		return fmt.Errorf("request Ollama chat response: %w", err)
	}

	content := strings.TrimSuffix(response.Content, "\n")

	if _, err := fmt.Fprintln(output, content); err != nil {
		return fmt.Errorf("write assistant response: %w", err)
	}

	return nil
}
