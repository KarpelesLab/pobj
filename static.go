package pobj

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/KarpelesLab/typutil"
)

type StaticMethod struct {
	fn       reflect.Value
	cnt      int            // number of actual args
	ctxPos   int            // pos of ctx argument, or -1
	arg      []reflect.Type // type used for the argument to the method
	variadic bool           // is the func's last argument a ...
}

var (
	ctxTyp = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Static returns a StaticMethod object for a func that accepts a context.Context and/or a
// struct object that is its arguments.
func Static(method any) *StaticMethod {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		panic("static method not a method")
	}

	typ := v.Type()
	res := &StaticMethod{fn: v, ctxPos: -1, cnt: typ.NumIn(), variadic: typ.IsVariadic()}

	ni := res.cnt

	for i := 0; i < ni; i += 1 {
		in := typ.In(i)
		if in.Implements(ctxTyp) {
			if res.ctxPos != -1 {
				panic("method taking multiple ctx arguments")
			}
			res.ctxPos = i
			res.cnt -= 1
			continue
		}
		res.arg = append(res.arg, in)
	}

	return res
}

// Call invokes the method without any argument. If the method expects some kind of argument, Call will attempt
// to get input_json from the context, and if obtained it will be parsed and passed as argument to the method.
func (s *StaticMethod) Call(ctx context.Context) (any, error) {
	// call this function, typically fetching request body from the context via input_json
	if s.cnt > 0 {
		// grab input json, call json.Unmarshal on argV
		input, ok := ctx.Value("input_json").(json.RawMessage)
		if ok {
			if s.cnt > 1 {
				// if the method take multiple arguments, the json value must be an array. By using RawMessage we
				// ensure we only parse the array part here, and not the contents, so it can be parsed for each
				// relevant argument type directly
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

// CallArg calls the method with the specified arguments. If less arguments are provided than required, an error will be raised.
func (s *StaticMethod) CallArg(ctx context.Context, arg ...any) (any, error) {
	if len(arg) < s.cnt {
		// not enough arguments to cover cnt
		return nil, ErrMissingArgs
	}
	if len(arg) > s.cnt && !s.variadic {
		return nil, errTooManyArgs
	}
	// call this function but pass arg values
	var args []reflect.Value
	var ctxPos int

	if s.ctxPos != -1 {
		args = make([]reflect.Value, len(arg)+1)
		args[s.ctxPos] = reflect.ValueOf(ctx)
		ctxPos = s.ctxPos
	} else {
		args = make([]reflect.Value, len(arg))
		ctxPos = len(arg) + 1
	}

	for argN, v := range arg {
		var argV reflect.Value
		if argN >= len(s.arg) {
			argV = reflect.New(s.arg[len(s.arg)-1])
		} else {
			argV = reflect.New(s.arg[argN])
		}
		err := typutil.AssignReflect(argV, reflect.ValueOf(v))
		if err != nil {
			return nil, err
		}

		if argN >= ctxPos {
			args[argN+1] = argV.Elem()
		} else {
			args[argN] = argV.Elem()
		}
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
