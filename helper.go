package pobj

import (
	"context"
	"fmt"
)

func (o *Object) ById(ctx context.Context, id string) (any, error) {
	if o.Action == nil {
		return nil, ErrMissingAction
	}
	get := o.Action.Fetch
	if get == nil {
		return nil, ErrMissingAction
	}
	return get.CallArg(ctx, struct{ Id string }{Id: id})
}

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
