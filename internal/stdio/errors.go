package stdio

import "errors"

var (
	// ErrIO indicates a standard input/output read or write failure.
	ErrIO = errors.New("failed to read or write data")

	// ErrInputEmpty indicates that no input was provided from stdin or arguments.
	ErrInputEmpty = errors.New("input is empty")
)
