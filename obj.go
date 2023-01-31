package pobj

import (
	"reflect"
	"strings"
)

type Object struct {
	name     string
	typ      reflect.Type
	children map[string]*Object
	static   map[string]*StaticMethod
	Action   *ObjectActions
	parent   *Object
}

// ObjectActions defines generic factories for usage in API calls
type ObjectActions struct {
	Fetch  *StaticMethod // Fetch action receives "id" and returns an instance (factory)
	List   *StaticMethod // List action returns a list of object
	Clear  *StaticMethod // Clear action deletes all objects and returns nothing
	Create *StaticMethod // Create action creates a new object and returns it
}

var (
	root = &Object{
		children: make(map[string]*Object),
	}
)

func lookup(p string, create bool) *Object {
	c := root

	pa := strings.Split(p, "/")

	for _, s := range pa {
		if c.children != nil {
			if x, ok := c.children[s]; ok {
				c = x
				continue
			}
		}
		if !create {
			return nil
		}
		if c.children == nil {
			c.children = make(map[string]*Object)
		}
		x := &Object{parent: c, name: s}
		c.children[s] = x
		c = x
	}
	return c
}

func Root() *Object {
	return root
}

func (o *Object) New() any {
	if o.typ == nil {
		return nil
	}
	return reflect.New(o.typ).Interface()
}

func (o *Object) String() string {
	switch o.parent {
	case root, nil:
		return o.name
	default:
		return o.parent.String() + "/" + o.name
	}
}

func (o *Object) Child(name string) *Object {
	if o == nil {
		return nil
	}
	if o.children == nil {
		return nil
	}
	res, ok := o.children[name]
	if !ok {
		return nil
	}
	return res
}

func (o *Object) Static(name string) *StaticMethod {
	if o == nil {
		return nil
	}
	if o.static == nil {
		return nil
	}
	res, ok := o.static[name]
	if !ok {
		return nil
	}
	return res
}
