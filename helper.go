package pobj

import (
	"context"
	"fmt"
)

// ById fetches an object instance by its ID using the object's Fetch action.
// This method requires the object to have a registered Fetch action.
// It automatically handles the appropriate argument passing format based on
// the Fetch action's signature.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: Unique identifier for the object to fetch
//
// Returns:
//   - The fetched object instance or an error if:
//   - No Action or Fetch action is registered
//   - The Fetch action fails
func (o *Object) ById(ctx context.Context, id string) (any, error) {
	if o.Action == nil {
		return nil, ErrMissingAction
	}
	get := o.Action.Fetch
	if get == nil {
		return nil, ErrMissingAction
	}
	if get.IsStringArg(0) {
		return get.CallArg(ctx, id)
	}
	return get.CallArg(ctx, struct{ Id string }{Id: id})
}

// ById is a generic helper that fetches a typed object by its ID.
// It automatically looks up the registered type, calls its Fetch action,
// and returns a properly typed result.
//
// Type parameter T should be the type you want to retrieve.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: Unique identifier for the object to fetch
//
// Returns:
//   - Pointer to the typed object or an error if:
//   - The type T is not registered
//   - The Fetch action fails
//   - The returned object is not of the expected type
func ById[T any](ctx context.Context, id string) (*T, error) {
	o := GetByType[T]()
	if o == nil {
		return nil, ErrUnknownType
	}
	res, err := o.ById(ctx, id)
	if err != nil {
		return nil, err
	}
	res_final, ok := res.(*T)
	if !ok {
		return nil, fmt.Errorf("pobj: bad type returned by Fetch, should have returned a %T but returned a %T", (*T)(nil), res)
	}
	return res_final, nil
}
