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
	static   map[string]*typutil.Callable // Static methods associated with this object (deprecated)
	methods  map[string]*Method           // Methods associated with this object
	fields   map[string]*Field            // Field metadata for struct types
	Action   *ObjectActions               // Actions that can be performed on this object type
	parent   *Object                      // Parent object in the hierarchy
	doc      string                       // Documentation for this object
}

// Field represents metadata about a struct field.
type Field struct {
	name   string       // Field name
	doc    string       // Documentation for this field
	typ    reflect.Type // Field type (from reflection)
	object *Object      // The object this field belongs to
}

// Method represents a registered method with its metadata.
// Methods can be either static (class-level) or require an instance in context.
type Method struct {
	callable         *typutil.Callable // The underlying callable function
	doc              string            // Documentation for this method
	requiresInstance bool              // If true, the object instance must be provided in context
	object           *Object           // The object this method belongs to
	name             string            // The method name
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
	// Check new methods map first
	if o.methods != nil {
		if m, ok := o.methods[name]; ok {
			return m.callable
		}
	}
	// Fall back to deprecated static map
	if o.static == nil {
		return nil
	}
	res, ok := o.static[name]
	if !ok {
		return nil
	}
	return res
}

// Method returns the registered method with the given name.
// Unlike Static, this returns the Method struct which includes metadata
// such as documentation and whether the method requires an instance.
// Returns nil if the method doesn't exist.
func (o *Object) Method(name string) *Method {
	if o == nil || o.methods == nil {
		return nil
	}
	return o.methods[name]
}

// Methods returns the names of all registered methods for this object.
// Returns nil if the object has no methods.
func (o *Object) Methods() []string {
	if o == nil || o.methods == nil {
		return nil
	}
	res := make([]string, 0, len(o.methods))
	for name := range o.methods {
		res = append(res, name)
	}
	return res
}

// SetDoc sets the documentation for this object and returns the object
// for method chaining.
func (o *Object) SetDoc(doc string) *Object {
	if o == nil {
		return nil
	}
	o.doc = doc
	return o
}

// Doc returns the documentation for this object.
func (o *Object) Doc() string {
	if o == nil {
		return ""
	}
	return o.doc
}

// Children returns the names of all direct child objects.
// Returns nil if the object has no children.
func (o *Object) Children() []string {
	if o == nil || o.children == nil {
		return nil
	}
	res := make([]string, 0, len(o.children))
	for name := range o.children {
		res = append(res, name)
	}
	return res
}

// All returns all registered Objects that have an associated type.
// This can be used for introspection and debugging.
func All() []*Object {
	mu.RLock()
	defer mu.RUnlock()
	res := make([]*Object, 0, len(typLookup))
	for _, o := range typLookup {
		res = append(res, o)
	}
	return res
}

// SetDoc sets the documentation for this method and returns the method
// for method chaining.
func (m *Method) SetDoc(doc string) *Method {
	if m == nil {
		return nil
	}
	m.doc = doc
	return m
}

// Doc returns the documentation for this method.
func (m *Method) Doc() string {
	if m == nil {
		return ""
	}
	return m.doc
}

// SetRequiresInstance marks this method as requiring an instance of the
// object to be provided in the context. Returns the method for chaining.
func (m *Method) SetRequiresInstance(requires bool) *Method {
	if m == nil {
		return nil
	}
	m.requiresInstance = requires
	return m
}

// RequiresInstance returns true if this method requires an instance of
// the object to be provided in the context when called.
func (m *Method) RequiresInstance() bool {
	if m == nil {
		return false
	}
	return m.requiresInstance
}

// Callable returns the underlying typutil.Callable for this method.
func (m *Method) Callable() *typutil.Callable {
	if m == nil {
		return nil
	}
	return m.callable
}

// Name returns the name of this method.
func (m *Method) Name() string {
	if m == nil {
		return ""
	}
	return m.name
}

// Object returns the Object this method belongs to.
func (m *Method) Object() *Object {
	if m == nil {
		return nil
	}
	return m.object
}

// String returns the full path of this method in "object:method" format.
func (m *Method) String() string {
	if m == nil {
		return ""
	}
	if m.object == nil {
		return m.name
	}
	return m.object.String() + ":" + m.name
}

// Field returns the field metadata for the given field name.
// Returns nil if the field doesn't exist or has no metadata.
func (o *Object) Field(name string) *Field {
	if o == nil || o.fields == nil {
		return nil
	}
	return o.fields[name]
}

// Fields returns the names of all fields with metadata.
// Returns nil if the object has no field metadata.
func (o *Object) Fields() []string {
	if o == nil || o.fields == nil {
		return nil
	}
	res := make([]string, 0, len(o.fields))
	for name := range o.fields {
		res = append(res, name)
	}
	return res
}

// SetFieldDoc sets the documentation for a field and returns the object for chaining.
// If the field doesn't exist in the metadata, it will be created.
func (o *Object) SetFieldDoc(fieldName, doc string) *Object {
	if o == nil {
		return nil
	}
	if o.fields == nil {
		o.fields = make(map[string]*Field)
	}
	f, ok := o.fields[fieldName]
	if !ok {
		f = &Field{
			name:   fieldName,
			object: o,
		}
		// Try to get the type from reflection
		if o.typ != nil && o.typ.Kind() == reflect.Struct {
			if sf, found := o.typ.FieldByName(fieldName); found {
				f.typ = sf.Type
			}
		}
		o.fields[fieldName] = f
	}
	f.doc = doc
	return o
}

// FieldDoc returns the documentation for a field.
// Returns empty string if the field has no documentation.
func (o *Object) FieldDoc(fieldName string) string {
	if o == nil || o.fields == nil {
		return ""
	}
	if f, ok := o.fields[fieldName]; ok {
		return f.doc
	}
	return ""
}

// SetDoc sets the documentation for this field and returns the field for chaining.
func (f *Field) SetDoc(doc string) *Field {
	if f == nil {
		return nil
	}
	f.doc = doc
	return f
}

// Doc returns the documentation for this field.
func (f *Field) Doc() string {
	if f == nil {
		return ""
	}
	return f.doc
}

// Name returns the name of this field.
func (f *Field) Name() string {
	if f == nil {
		return ""
	}
	return f.name
}

// Type returns the reflect.Type of this field.
func (f *Field) Type() reflect.Type {
	if f == nil {
		return nil
	}
	return f.typ
}

// Object returns the Object this field belongs to.
func (f *Field) Object() *Object {
	if f == nil {
		return nil
	}
	return f.object
}
