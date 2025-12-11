[![GoDoc](https://godoc.org/github.com/KarpelesLab/pobj?status.svg)](https://godoc.org/github.com/KarpelesLab/pobj)

# pobj

A Go object registry library that provides hierarchical type management, allowing types to be registered, instantiated by name, and accessed through a tree-like structure. It supports static methods, REST-like actions, and type-safe generic lookups.

## Features

- **Hierarchical Registry** - Organize types in a tree structure using path-based names (e.g., `user/admin`)
- **Type-Safe Generics** - Uses Go 1.18+ generics for compile-time type safety
- **REST-like Actions** - Built-in support for Fetch, List, Create, and Clear operations
- **Static Methods** - Register type-level functions that can be called by name
- **Reflection-Based Instantiation** - Create new instances of registered types at runtime
- **Context Support** - All operations support `context.Context` for cancellation, timeouts, and passing contextual information (e.g., database connections, request metadata)

## Installation

```bash
go get github.com/KarpelesLab/pobj
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/KarpelesLab/pobj"
    "github.com/KarpelesLab/typutil"
)

// Define your type
type User struct {
    ID    string
    Name  string
    Email string
}

func main() {
    // Register the type
    pobj.Register[User]("user")

    // Create a new instance
    obj := pobj.Get("user")
    instance := obj.New().(*User)
    instance.Name = "John Doe"

    fmt.Printf("Created user: %+v\n", instance)
}
```

## Usage

### Registering Types

Register a type with a path name:

```go
// Simple registration
pobj.Register[User]("user")

// Nested registration (creates hierarchy)
pobj.Register[AdminUser]("user/admin")
pobj.Register[GuestUser]("user/guest")
```

### Registering Actions

Register REST-like actions for a type:

```go
actions := &pobj.ObjectActions{
    Fetch: typutil.Func(func(ctx context.Context, id string) (*User, error) {
        return fetchUserFromDB(id)
    }),
    List: typutil.Func(func(ctx context.Context) ([]*User, error) {
        return listUsersFromDB()
    }),
    Create: typutil.Func(func(ctx context.Context, data *User) (*User, error) {
        return createUserInDB(data)
    }),
    Clear: typutil.Func(func(ctx context.Context) error {
        return clearAllUsers()
    }),
}

pobj.RegisterActions[User]("user", actions)
```

### Registering Static Methods

Register functions associated with a type (not instance methods):

```go
// Format: "path:methodName"
pobj.RegisterStatic("user:getByEmail", func(ctx context.Context, email string) (*User, error) {
    return getUserByEmail(email)
})

// Call the static method
obj := pobj.Get("user")
method := obj.Static("getByEmail")
result, err := typutil.Call[*User](method, ctx, "user@example.com")
```

### Retrieving Objects

```go
// By path name
obj := pobj.Get("user")
obj = pobj.Get("user/admin")  // nested path

// By type (generic)
obj := pobj.GetByType[User]()

// Navigate hierarchy
root := pobj.Root()
child := root.Child("user")
```

### Creating Instances

```go
obj := pobj.Get("user")
instance := obj.New()  // returns any, type assert as needed

user := instance.(*User)
```

### Fetching by ID

Using the registered Fetch action:

```go
ctx := context.Background()

// Method 1: Via Object
obj := pobj.Get("user")
result, err := obj.ById(ctx, "user-123")
user := result.(*User)

// Method 2: Generic helper (type-safe)
user, err := pobj.ById[User](ctx, "user-123")
```

## API Reference

### Core Types

#### Object

Represents a registered type in the registry:

```go
type Object struct {
    Action *ObjectActions  // REST-like actions (exported)
    // ... internal fields
}
```

Methods:
- `New() any` - Create a new instance of the registered type
- `String() string` - Get the full path name
- `Child(name string) *Object` - Get a direct child object
- `Static(name string) *typutil.Callable` - Get a registered static method
- `ById(ctx context.Context, id string) (any, error)` - Fetch instance by ID

#### ObjectActions

Defines factory functions for API operations:

```go
type ObjectActions struct {
    Fetch  *typutil.Callable  // Get object by ID
    List   *typutil.Callable  // List all objects
    Create *typutil.Callable  // Create new object
    Clear  *typutil.Callable  // Delete all objects
}
```

### Functions

| Function | Description |
|----------|-------------|
| `Register[T any](name string) *Object` | Register a type with a path name |
| `RegisterActions[T any](name string, actions *ObjectActions)` | Register a type with actions |
| `RegisterStatic(name string, fn any)` | Register a static method (`path:method` format) |
| `Get(name string) *Object` | Get object by path (returns nil if not found) |
| `GetByType[T any]() *Object` | Get object by generic type |
| `Root() *Object` | Get the root of the hierarchy |
| `ById[T any](ctx, id string) (*T, error)` | Type-safe fetch by ID |

### Errors

| Error | Description |
|-------|-------------|
| `ErrUnknownType` | Type is not registered |
| `ErrMissingAction` | Required action (e.g., Fetch) is not registered |

## Fetch Argument Format

The Fetch action supports two argument formats:

1. **String argument** (recommended):
   ```go
   func(ctx context.Context, id string) (*User, error)
   ```

2. **Struct argument** (legacy):
   ```go
   func(ctx context.Context, args struct{ Id string }) (*User, error)
   ```

The library automatically detects which format your Fetch function uses.

## Panic Behavior

The following operations will panic:

- Registering the same path twice with different types
- Using invalid static method name format (missing `:`)
- Passing a non-function to `RegisterStatic`

## Dependencies

- [github.com/KarpelesLab/typutil](https://github.com/KarpelesLab/typutil) - Type utilities and callable wrappers

## License

See [LICENSE](LICENSE) file.
