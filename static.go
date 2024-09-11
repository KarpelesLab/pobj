package pobj

import (
	"context"
	"encoding"
	"encoding/json"
	"reflect"

	"github.com/KarpelesLab/typutil"
)

type StaticMethod struct {
	fn     reflect.Value
	cnt    int            // number of actual args
	ctxPos int            // pos of ctx argument, or -1
	argPos []int          // pos of arguments, or nil
	arg    []reflect.Type // type used for the argument to the method
}

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	ctxTyp              = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Static returns a StaticMethod object for a func that accepts a context.Context and/or a
// struct object that is its arguments.
func Static(method any) *StaticMethod {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		panic("static method not a method")
	}

	typ := v.Type()
	res := &StaticMethod{fn: v, ctxPos: -1, cnt: typ.NumIn()}

	ni := res.cnt

	for i := 0; i < ni; i += 1 {
		in := typ.In(i)
		if in.Implements(ctxTyp) {
			if res.ctxPos != -1 {
				panic("method taking multiple ctx arguments")
			}
			res.ctxPos = i
			continue
		}
		res.argPos = append(res.argPos, i)
		res.arg = append(res.arg, in)
	}

	return res
}

func (s *StaticMethod) Call(ctx context.Context) (any, error) {
	// call this function, typically fetching request body from the context via input_json
	if len(s.argPos) > 0 {
		// grab input json, call json.Unmarshal on argV
		input, ok := ctx.Value("input_json").(json.RawMessage)
		if ok {
			if len(s.argPos) > 1 {
				var args []typutil.RawJsonMessage
				err := json.Unmarshal(input, &args)
				if err != nil {
					return nil, err
				}
				anyArgs := make([]any, len(args))
				for n, v := range args {
					anyArgs[n] = v
				}
				return s.CallArg(ctx, anyArgs...)
			}
			return s.CallArg(ctx, typutil.RawJsonMessage(input))
		}
	}

	return s.CallArg(ctx)
}

func (s *StaticMethod) CallArg(ctx context.Context, arg ...any) (any, error) {
	// call this function but pass arg values
	args := make([]reflect.Value, s.cnt)
	if s.ctxPos != -1 {
		args[s.ctxPos] = reflect.ValueOf(ctx)
	}
	for argN, pos := range s.argPos {
		argV := reflect.New(s.arg[argN])
		err := typutil.AssignReflect(argV, reflect.ValueOf(arg[argN]))
		if err != nil {
			return nil, err
		}

		args[pos] = argV.Elem()
	}

	return s.parseResult(s.fn.Call(args))
}

func (s *StaticMethod) IsStringArg(n int) bool {
	return s.ArgKind(n) == reflect.String
}

func (s *StaticMethod) ArgKind(n int) reflect.Kind {
	if n >= len(s.arg) {
		return reflect.Invalid
	}
	return s.arg[n].Kind()
}

var errTyp = reflect.TypeOf((*error)(nil)).Elem()

func (s *StaticMethod) parseResult(res []reflect.Value) (output any, err error) {
	// for each value in res, try to find which one is an error and which one is a result
	for _, v := range res {
		if v.Type().Implements(errTyp) {
			err, _ = v.Interface().(error)
			continue
		}
		output = v.Interface()
	}
	return
}

func Call[T any](s *StaticMethod, ctx context.Context, arg ...any) (T, error) {
	res, err := s.CallArg(ctx, arg...)
	if v, ok := res.(T); ok {
		return v, err
	}
	return reflect.New(reflect.TypeFor[T]()).Elem().Interface().(T), nil
}
