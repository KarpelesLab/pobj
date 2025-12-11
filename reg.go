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
func RegisterStatic(name string, fn any) {
	pos := strings.IndexByte(name, ':')
	if pos == -1 {
		panic(fmt.Sprintf("invalid name %s for static method", name))
	}

	static := typutil.Func(fn)
	if static == nil {
		panic(fmt.Sprintf("invalid static method %T", fn))
	}

	mu.Lock()
	defer mu.Unlock()
	o := lookup(name[:pos], true)
	name = name[pos+1:]

	if o.static == nil {
		o.static = make(map[string]*typutil.Callable)
	}

	o.static[name] = static
}

// RegisterActions registers a type with associated actions for API operations.
// The actions include common operations like Fetch, List, Clear, and Create.
// Similar to Register, but also associates the ObjectActions with the registered type.
// Intended for implementing REST-like operations on the registered type.
// Panics if the name is already registered with a different type.
func RegisterActions[T any](name string, actions *ObjectActions) {
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
}
