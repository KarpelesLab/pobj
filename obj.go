// Package pobj provides an object registry system for Go, allowing types to be
// registered, instantiated by name, and accessed through a hierarchical structure.
// It supports static methods, object actions, and type-based lookup.
package pobj

import (
	"reflect"
	"strings"
	"sync"

	"github.com/KarpelesLab/typutil"
)

// Object represents a registered type in the object registry.
// Objects can be organized hierarchically with parent/child relationships.
type Object struct {
	name     string                       // Name of the object in the registry
	typ      reflect.Type                 // The Go type represented by this object
	children map[string]*Object           // Child objects in the hierarchy (name → object)
	static   map[string]*typutil.Callable // Static methods associated with this object
	Action   *ObjectActions               // Actions that can be performed on this object type
	parent   *Object                      // Parent object in the hierarchy
}

// ObjectActions defines callable factories for REST-like API operations.
// Each action is optional and can be set to nil if not needed.
type ObjectActions struct {
	Fetch  *typutil.Callable // Fetch retrieves a single object by ID
	List   *typutil.Callable // List returns all objects of this type
	Clear  *typutil.Callable // Clear deletes all objects of this type
	Create *typutil.Callable // Create instantiates a new object
}

var (
	// root is the top-level object in the hierarchy
	root = &Object{
		children: make(map[string]*Object),
	}
	// typLookup provides direct access to objects by their reflected type
	typLookup = make(map[reflect.Type]*Object)
	// mu protects access to root and typLookup
	mu sync.RWMutex
)

// lookup finds an Object by its path in the hierarchy.
// If create is true, it will create missing objects along the path.
// Paths use '/' as a separator, e.g. "user/admin" to locate nested objects.
// Caller must hold appropriate lock (read lock if create=false, write lock if create=true).
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

// Root returns the root object holder, which is the top-level object
// in the hierarchical registry.
func Root() *Object {
	mu.RLock()
	defer mu.RUnlock()
	return root
}

// Get returns the Object matching the given name, or nil if no such object exists.
// The name can be a path using '/' as separator for nested objects.
func Get(name string) *Object {
	mu.RLock()
	defer mu.RUnlock()
	return lookup(name, false)
}

// GetByType returns the Object matching the given generic type parameter.
// It handles pointer types by unwrapping them to their underlying type.
// Returns nil if the type is not registered.
func GetByType[T any]() *Object {
	t := reflect.TypeOf((*T)(nil))
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	mu.RLock()
	defer mu.RUnlock()
	if o, ok := typLookup[t]; ok {
		return o
	}
	return nil
}

// New creates and returns a new instance of the registered type.
// Returns nil if the Object doesn't have an associated type.
// The returned value will be a pointer to a newly allocated instance.
func (o *Object) New() any {
	if o.typ == nil {
		return nil
	}
	return reflect.New(o.typ).Interface()
}

// String returns the full path name of this Object in the registry hierarchy.
// The path uses '/' as a separator between parent and child objects.
func (o *Object) String() string {
	switch o.parent {
	case root, nil:
		return o.name
	default:
		return o.parent.String() + "/" + o.name
	}
}

// Child retrieves a direct child Object with the given name.
// Returns nil if the object has no children or the requested child doesn't exist.
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

// Static returns the registered static method with the given name.
// Static methods are functions associated with an object type rather than
// with a specific instance of that type.
// Returns nil if the object has no static methods or the requested method doesn't exist.
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
