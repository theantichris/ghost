package app

import "errors"

type App struct {
	model  string
	stream bool
}

func New(model string, stream bool) (*App, error) {
	if model == "" {
		return nil, errors.New("model cannot be empty")
	}

	app := &App{
		model:  model,
		stream: stream,
	}

	return app, nil
}
