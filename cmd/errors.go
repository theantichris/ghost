package cmd

import "errors"

var (
	ErrConfig     = errors.New("failed to bind config")
	ErrLogging    = errors.New("failed to setup logging")
	ErrInput      = errors.New("failed to get input")
	ErrLLM        = errors.New("failed to process LLM request")
	ErrIO         = errors.New("failed to read or write data")
	ErrInputEmpty = errors.New("input is empty")
)
