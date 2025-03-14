package pobj

import (
	"context"

	"github.com/KarpelesLab/typutil"
)

// Static returns a [typutil.Callable] object for a function that accepts a context.Context and/or a
// struct object as its arguments. This facilitates calling functions with dynamic arguments.
//
// The function should follow the conventions expected by typutil.Func.
//
// Deprecated: use [typutil.Func] directly instead. This function is maintained for backward
// compatibility but will be removed in a future version.
func Static(method any) *typutil.Callable {
	return typutil.Func(method)
}

// Call calls the provided method and converts the result to the specified type T.
// This is a generic helper that handles type conversion of the return value.
//
// Parameters:
//   - s: The callable to invoke
//   - ctx: Context for the operation
//   - arg: Variable arguments to pass to the callable
//
// Returns:
//   - A value of type T and an error (if any)
//
// Deprecated: use [typutil.Call] directly instead. This function is maintained for backward
// compatibility but will be removed in a future version.
func Call[T any](s *typutil.Callable, ctx context.Context, arg ...any) (T, error) {
	return typutil.Call[T](s, ctx, arg...)
}
