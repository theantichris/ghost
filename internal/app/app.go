package app

import (
	"errors"
	"fmt"
)

type App struct {
	model  string
	stream bool
}

func New(model string, stream bool) (*App, error) {
	if model == "" {
		return nil, fmt.Errorf("app init: %w", errors.New("model cannot be empty"))
	}

	app := &App{
		model:  model,
		stream: stream,
	}

	return app, nil
}
