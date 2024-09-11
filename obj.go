package pobj

import (
	"reflect"
	"strings"

	"github.com/KarpelesLab/typutil"
)

type Object struct {
	name     string
	typ      reflect.Type
	children map[string]*Object
	static   map[string]*typutil.Callable
	Action   *ObjectActions
	parent   *Object
}

// ObjectActions defines generic factories for usage in API calls
type ObjectActions struct {
	Fetch  *typutil.Callable // Fetch action receives "id" and returns an instance (factory)
	List   *typutil.Callable // List action returns a list of object
	Clear  *typutil.Callable // Clear action deletes all objects and returns nothing
	Create *typutil.Callable // Create action creates a new object and returns it
}

var (
	root = &Object{
		children: make(map[string]*Object),
	}
	typLookup = make(map[reflect.Type]*Object)
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

// Root returns the root object holder
func Root() *Object {
	return root
}

// Get returns the Object matching the given name, or nil if no such object exists
func Get(name string) *Object {
	return lookup(name, false)
}

// GetByType returns the Object matching the type given on the command line
func GetByType[T any]() *Object {
	t := reflect.TypeOf((*T)(nil))
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if o, ok := typLookup[t]; ok {
		return o
	}
	return nil
}

// New returns a new instance of the given object
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

// Static returns the given static method
func (o *Object) Static(name string) *typutil.Callable {
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
