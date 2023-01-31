package pobj

import (
	"fmt"
	"reflect"
	"strings"
)

// Register adds the given object to the registry of name-instanciable objects
func Register(name string, obj any) *Object {
	o := lookup(name, true)
	if o.typ != nil {
		panic(fmt.Sprintf("multiple registrations for type %s (%T), existing = %+v", name, obj, o))
	}
	o.typ = reflect.TypeOf(obj)
	return o
}

// RegisterStatic adds a static method to an object
func RegisterStatic(name string, fn any) {
	pos := strings.IndexByte(name, ':')
	if pos == -1 {
		panic(fmt.Sprintf("invalid name %s for static method", name))
	}

	static := Static(fn)
	if static == nil {
		panic(fmt.Sprintf("invalid static method %T", fn))
	}

	o := lookup(name[:pos], true)
	name = name[pos+1:]

	if o.static == nil {
		o.static = make(map[string]*StaticMethod)
	}

	o.static[name] = static
}

// RegisterActions is used for static REST methods such as get (factory) and
// list. Methods such as update and delete require an object.
func RegisterActions(name string, actions *ObjectActions) {
	o := lookup(name, true)
	o.Action = actions
}
