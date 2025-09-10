package llm

import "errors"

var ErrURLEmpty = errors.New("baseURL cannot be empty")
var ErrModelEmpty = errors.New("defaultModel cannot be empty")
