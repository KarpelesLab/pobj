package pobj

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/KarpelesLab/typutil"
)

// Register adds a type to the registry with the given name.
// The type T is determined by the generic parameter.
// Name can be a path using '/' as separator for nested object registration.
// Returns the registered Object for further configuration.
// Panics if the name is already registered with a different type.
func Register[T any](name string) *Object {
	mu.Lock()
	defer mu.Unlock()
	o := lookup(name, true)
	if o.typ != nil {
		panic(fmt.Sprintf("multiple registrations for type %s (%T), existing = %+v", name, (*T)(nil), o))
	}
	o.typ = reflect.TypeOf((*T)(nil))
	for o.typ.Kind() == reflect.Pointer {
		o.typ = o.typ.Elem()
	}
	typLookup[o.typ] = o
	return o
}

// RegisterStatic adds a static method to an object.
// The name must be in the format "object/path:methodName" where:
// - "object/path" is the registered object's path
// - "methodName" is the name of the static method
// The function fn will be converted to a callable using typutil.Func.
// Panics if the name format is invalid or the function cannot be converted.
// Deprecated: Use RegisterMethod instead which returns *Method for chaining.
func RegisterStatic(name string, fn any) {
	RegisterMethod(name, fn)
}

// RegisterMethod adds a method to an object and returns the Method for further configuration.
// The name must be in the format "object/path:methodName" where:
// - "object/path" is the registered object's path
// - "methodName" is the name of the method
// The function fn will be converted to a callable using typutil.Func.
// Panics if the name format is invalid or the function cannot be converted.
//
// The returned Method can be used to set documentation and other properties:
//
//	pobj.RegisterMethod("User:getByEmail", getByEmail).
//	    SetDoc("Fetch a user by their email address").
//	    SetRequiresInstance(false)
func RegisterMethod(name string, fn any) *Method {
	pos := strings.IndexByte(name, ':')
	if pos == -1 {
		panic(fmt.Sprintf("invalid name %s for method", name))
	}

	callable := typutil.Func(fn)
	if callable == nil {
		panic(fmt.Sprintf("invalid method %T", fn))
	}

	mu.Lock()
	defer mu.Unlock()
	o := lookup(name[:pos], true)
	methodName := name[pos+1:]

	if o.methods == nil {
		o.methods = make(map[string]*Method)
	}

	m := &Method{
		callable: callable,
		object:   o,
		name:     methodName,
	}
	o.methods[methodName] = m
	return m
}

// RegisterActions registers a type with associated actions for API operations.
// The actions include common operations like Fetch, List, Clear, and Create.
// Similar to Register, but also associates the ObjectActions with the registered type.
// Intended for implementing REST-like operations on the registered type.
// Returns the registered Object for further configuration.
// Panics if the name is already registered with a different type.
func RegisterActions[T any](name string, actions *ObjectActions) *Object {
	mu.Lock()
	defer mu.Unlock()
	o := lookup(name, true)
	if o.typ != nil {
		panic(fmt.Sprintf("multiple registrations for type %s (%T), existing = %+v", name, (*T)(nil), o))
	}
	o.typ = reflect.TypeOf((*T)(nil))
	for o.typ.Kind() == reflect.Pointer {
		o.typ = o.typ.Elem()
	}
	typLookup[o.typ] = o
	o.Action = actions
	return o
}
