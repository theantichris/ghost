package llm

import "errors"

var (
	ErrNoBaseURL      = errors.New("no base URL provided")
	ErrNoDefaultModel = errors.New("no default model provided")
)
