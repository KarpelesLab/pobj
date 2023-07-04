package pobj

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
)

type StaticMethod struct {
	fn     reflect.Value
	cnt    int          // number of actual args
	ctxPos int          // pos of ctx argument, or -1
	argPos int          // pos of arguments, or -1
	arg    reflect.Type // type used for the argument to the method
	argPtr bool         // is argument a ptr?
}

type valueScanner interface {
	Scan(any) error
}

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	valueScannerType    = reflect.TypeOf((*valueScanner)(nil)).Elem()
)

// Static returns a StaticMethod object for a func that accepts a context.Context and/or a
// struct object that is its arguments.
func Static(method any) *StaticMethod {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		panic("static method not a method")
	}

	typ := v.Type()
	res := &StaticMethod{fn: v, ctxPos: -1, argPos: -1, cnt: typ.NumIn()}

	ni := res.cnt
	ctxTyp := reflect.TypeOf((*context.Context)(nil)).Elem()

	for i := 0; i < ni; i += 1 {
		in := typ.In(i)
		if in.Implements(ctxTyp) {
			if res.ctxPos != -1 {
				panic("method taking multiple ctx arguments")
			}
			res.ctxPos = i
			continue
		}
		if in.Kind() == reflect.Ptr {
			in = in.Elem()
			res.argPtr = true
		}
		if in.Kind() == reflect.Struct {
			if res.argPos != -1 {
				panic("method taking multiple arg arguments")
			}
			res.argPos = i
			res.arg = in
			continue
		}
		panic(fmt.Sprintf("method has unknown type of argument %s", in))
	}

	return res
}

func (s *StaticMethod) Call(ctx context.Context) (any, error) {
	// call this function
	args := make([]reflect.Value, s.cnt)

	if s.ctxPos != -1 {
		args[s.ctxPos] = reflect.ValueOf(ctx)
	}
	if s.argPos != -1 {
		argV := reflect.New(s.arg)

		// grab input json, call json.Unmarshal on argV
		input, ok := ctx.Value("input_json").(json.RawMessage)
		if ok {
			err := json.Unmarshal(input, argV.Interface())
			if err != nil {
				return nil, err
			}
		}

		if !s.argPtr {
			argV = argV.Elem()
		}
		args[s.argPos] = argV
	}

	return s.parseResult(s.fn.Call(args))
}

func (s *StaticMethod) CallArg(ctx context.Context, arg any) (any, error) {
	// call this function but pass arg values
	args := make([]reflect.Value, s.cnt)
	if s.ctxPos != -1 {
		args[s.ctxPos] = reflect.ValueOf(ctx)
	}
	if s.argPos != -1 {
		argV := reflect.New(s.arg)
		argVE := argV.Elem()

		// index location of values
		argIn := reflect.ValueOf(arg)
		if argIn.Kind() == reflect.Ptr {
			if argIn.IsNil() {
				return s.parseResult(s.fn.Call(args))
			}
			argIn = argIn.Elem()
		}
		if argIn.Kind() == reflect.Interface {
			if argIn.IsNil() {
				return s.parseResult(s.fn.Call(args))
			}
		}

		switch argIn.Kind() {
		case reflect.Struct:
			cnt := argIn.Type().NumField()
			for i := 0; i < cnt; i++ {
				inFld := argIn.Type().Field(i)
				outFld, ok := argVE.Type().FieldByName(inFld.Name)
				if !ok {
					// ignore field... ?
					continue
				}
				err := assignValueTo(argVE.Field(outFld.Index[0]), argIn.Field(i))
				if err != nil {
					return nil, err
				}
			}
		case reflect.Map:
			if argIn.Type().Key().Kind() != reflect.String {
				return nil, fmt.Errorf("map key must be a string")
			}
			iter := argIn.MapRange()
			for iter.Next() {
				k := iter.Key()
				outFld, ok := argVE.Type().FieldByName(k.String())
				if !ok {
					// ignore field
					continue
				}
				err := assignValueTo(argVE.Field(outFld.Index[0]), iter.Value())
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("arg must be a struct, %T passed", arg)
		}
		if !s.argPtr {
			argV = argV.Elem()
		}
		args[s.argPos] = argV
	}

	return s.parseResult(s.fn.Call(args))
}

func assignValueTo(dst reflect.Value, src reflect.Value) error {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
	}
	dstt := dst.Type()
	srct := src.Type()

	if srct.AssignableTo(dstt) {
		dst.Set(src)
		return nil
	}
	if srct.ConvertibleTo(dstt) {
		dst.Set(src.Convert(dstt))
		return nil
	}
	if dstt.Kind() == reflect.Pointer && dstt.Implements(valueScannerType) {
		if dst.IsNil() {
			dst.Set(reflect.New(dstt.Elem()))
		}
		return dst.Interface().(valueScanner).Scan(src.Interface())
	} else if dst.CanAddr() && reflect.PointerTo(dstt).Implements(valueScannerType) {
		return dst.Addr().Interface().(valueScanner).Scan(src.Interface())
	}
	return fmt.Errorf("incompatible source field type %#v assign to %#v", srct.Name(), dstt.Name())
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
