package pobj

import (
	"context"

	"github.com/KarpelesLab/typutil"
)

// Static returns a [typutil.Callable] object for a func that accepts a context.Context and/or a
// struct object that is its arguments.
//
// Deprecated: use [typutil.Func] instead
func Static(method any) *typutil.Callable {
	return typutil.Func(method)
}

// Call calls the provided method and converts the result to the specified type
//
// Deprecated: use [typutil.Call] instead
func Call[T any](s *typutil.Callable, ctx context.Context, arg ...any) (T, error) {
	return typutil.Call[T](s, ctx, arg...)
}
