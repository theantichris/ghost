package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/theantichris/ghost/v4/internal/conversation"
	"github.com/theantichris/ghost/v4/internal/ollama"
)

var (
	errModelRequired  = errors.New("model is required: use --model <name>")
	errPromptRequired = errors.New("prompt is required")
)

func main() {
	client, err := ollama.NewClient(ollama.DefaultURL, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ghost // uplink initialization failed: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, client.ChatStream); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ghost // uplink failure: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer, chat conversation.ChatStreamFunc) error {
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

	session := conversation.NewSession()

	emit := func(event conversation.TurnEvent) error {
		var content string

		switch event.Kind {
		case conversation.TurnEventContent:
			content = event.Content
		case conversation.TurnEventCompleted:
			if strings.HasSuffix(event.Content, "\n") {
				return nil
			}

			content = "\n"
		default:
			return nil
		}

		if _, err := io.WriteString(output, content); err != nil {
			return fmt.Errorf("write assistant response: %w", err)
		}

		return nil
	}

	_, err := conversation.RunTurnStream(ctx, session, 1, modelName, prompt, chat, emit)
	if err != nil {
		return fmt.Errorf("request Ollama chat response: %w", err)
	}

	return nil
}
