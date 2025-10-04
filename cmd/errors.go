package cmd

import "errors"

var (
	ErrConfig  = errors.New("failed to bind config")
	ErrLogging = errors.New("failed to setup logging")
	ErrLLM     = errors.New("failed to process LLM request")
)
