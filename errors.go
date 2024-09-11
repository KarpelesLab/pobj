package pobj

import "errors"

var (
	ErrUnknownType   = errors.New("pobj: unknown object type")
	ErrMissingAction = errors.New("pobj: no such action exists")
	ErrMissingArgs   = errors.New("pobj: missing arguments")
	errTooManyArgs   = errors.New("pobj: too many arguments")
)
