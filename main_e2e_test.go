//go:build e2e

package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/theantichris/ghost/v4/internal/ollama"
)

const defaultE2EModel = "qwen3:0.6b"

func TestOneShotWorkflow(t *testing.T) {
	model := strings.TrimSpace(os.Getenv("GHOST_E2E_MODEL"))
	if model == "" {
		model = defaultE2EModel
	}

	client, err := ollama.NewClient(ollama.DefaultURL, nil)
	if err != nil {
		t.Fatalf("NewClient() error = %v, want nil", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var output bytes.Buffer

	err = run(ctx, []string{"--model", model, "Reply with one short sentence confirming Ghost is online."}, &output, client.Chat)
	if err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}

	response := strings.TrimSpace(output.String())
	if response == "" {
		t.Fatal("run() output is blank, want a nonblank Ollama response")
	}

	t.Logf("model: %s", model)
	t.Logf("Ghost response: %s", response)
}
