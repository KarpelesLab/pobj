package pobj

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/KarpelesLab/typutil"
)

// Register adds the given object to the registry of name-instanciable objects
func Register[T any](name string) *Object {
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

// RegisterStatic adds a static method to an object
func RegisterStatic(name string, fn any) {
	pos := strings.IndexByte(name, ':')
	if pos == -1 {
		panic(fmt.Sprintf("invalid name %s for static method", name))
	}

	static := typutil.Func(fn)
	if static == nil {
		panic(fmt.Sprintf("invalid static method %T", fn))
	}

	o := lookup(name[:pos], true)
	name = name[pos+1:]

	if o.static == nil {
		o.static = make(map[string]*typutil.Callable)
	}

	o.static[name] = static
}

// RegisterActions is used for static REST methods such as get (factory) and
// list. Methods such as update and delete require an object.
func RegisterActions[T any](name string, actions *ObjectActions) {
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
