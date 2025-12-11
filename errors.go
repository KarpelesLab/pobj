package pobj

import "errors"

// Sentinel errors for common failure cases.
var (
	// ErrUnknownType is returned when trying to get an object by a type
	// that hasn't been registered with the package. This typically occurs
	// when using GetByType[T]() or ById[T]() with an unregistered type.
	ErrUnknownType = errors.New("pobj: unknown object type")

	// ErrMissingAction is returned when trying to use an action (like Fetch)
	// that hasn't been registered for the object. This happens when an object
	// has no associated ObjectActions or when the specific action being used
	// is nil within the ObjectActions.
	ErrMissingAction = errors.New("pobj: no such action exists")
)
