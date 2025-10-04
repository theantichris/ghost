package stdio

import "errors"

var (
	ErrIO         = errors.New("failed to read or write data")
	ErrInputEmpty = errors.New("input is empty")
)
