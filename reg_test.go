package pobj_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/pobj"
	"github.com/KarpelesLab/typutil"
)

func TestRegister(t *testing.T) {
	// Test that Register returns the correct object
	obj := pobj.Register[TestPerson]("test/register/person")
	if obj == nil {
		t.Fatal("Register returned nil")
	}

	// Verify that the object can be retrieved
	retrieved := pobj.Get("test/register/person")
	if retrieved == nil {
		t.Fatal("Failed to get registered object")
	}

	// Verify that the object can be retrieved by type
	typeRetrieved := pobj.GetByType[TestPerson]()
	if typeRetrieved == nil {
		t.Fatal("Failed to get object by type")
	}

	// Make sure type lookup works
	instance := retrieved.New()
	if instance == nil {
		t.Fatal("Failed to create new instance")
	}

	_, ok := instance.(*TestPerson)
	if !ok {
		t.Errorf("Wrong type, expected *TestPerson, got %T", instance)
	}
}

func TestRegisterActions(t *testing.T) {
	actions := &pobj.ObjectActions{
		Fetch: typutil.Func(func(ctx context.Context, id string) (*TestPerson, error) {
			return &TestPerson{ID: id, Name: "Test Action Person"}, nil
		}),
	}

	// Register with actions
	pobj.RegisterActions[TestPerson]("test/register/with-actions", actions)

	// Get the registered object
	obj := pobj.Get("test/register/with-actions")
	if obj == nil {
		t.Fatal("Failed to get registered object")
	}

	// Verify that the object has the actions
	if obj.Action == nil {
		t.Fatal("Object has no actions")
	}

	if obj.Action.Fetch == nil {
		t.Fatal("Object has no Fetch action")
	}

	// Test the fetch action
	ctx := context.Background()
	result, err := obj.ById(ctx, "action-test-id")
	if err != nil {
		t.Fatalf("Failed to fetch by ID: %v", err)
	}

	person, ok := result.(*TestPerson)
	if !ok {
		t.Fatalf("Wrong type, got %T, want *TestPerson", result)
	}

	if person.ID != "action-test-id" {
		t.Errorf("Wrong ID, got %s, want %s", person.ID, "action-test-id")
	}

	if person.Name != "Test Action Person" {
		t.Errorf("Wrong name, got %s, want %s", person.Name, "Test Action Person")
	}
}

func TestRegisterStatic(t *testing.T) {
	// Register a static method
	pobj.RegisterStatic("test/static-test:method", func(ctx context.Context) (string, error) {
		return "static test result", nil
	})

	// Get the object
	obj := pobj.Get("test/static-test")
	if obj == nil {
		t.Fatal("Failed to get object with static method")
	}

	// Get the static method
	method := obj.Static("method")
	if method == nil {
		t.Fatal("Failed to get static method")
	}

	// Call the static method
	result, err := typutil.Call[string](method, context.Background())
	if err != nil {
		t.Fatalf("Failed to call static method: %v", err)
	}

	if result != "static test result" {
		t.Errorf("Wrong result, got %s, want %s", result, "static test result")
	}
}

// TestPanicCases tests cases that should panic
// These tests are separate because they cause panics
func testRegisterDuplicate(t *testing.T) {
	// This function is called from TestPanicRecovery
	// Register a type
	pobj.Register[TestPerson]("test/panic/duplicate")

	// Register the same path with a different type - should panic
	pobj.Register[TestCompany]("test/panic/duplicate")
}

func testRegisterActionsDuplicate(t *testing.T) {
	// This function is called from TestPanicRecovery
	actions := &pobj.ObjectActions{}

	// Register a type with actions
	pobj.RegisterActions[TestPerson]("test/panic/actions-duplicate", actions)

	// Register the same path with a different type - should panic
	pobj.RegisterActions[TestCompany]("test/panic/actions-duplicate", actions)
}

func testRegisterStaticInvalidName(t *testing.T) {
	// This function is called from TestPanicRecovery
	// Register a static method with invalid name (no colon) - should panic
	pobj.RegisterStatic("invalid-name-no-colon", func() {})
}

func testRegisterStaticInvalidFunction(t *testing.T) {
	// This function is called from TestPanicRecovery
	// Register a static method with invalid function - should panic
	pobj.RegisterStatic("test:invalid", "not a function")
}

func TestPanicRecovery(t *testing.T) {
	// Test duplicate registration - should panic
	t.Run("Register duplicate", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic for duplicate registration, but no panic occurred")
			}
		}()

		testRegisterDuplicate(t)
	})

	// Test duplicate registration with actions - should panic
	t.Run("RegisterActions duplicate", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic for duplicate registration with actions, but no panic occurred")
			}
		}()

		testRegisterActionsDuplicate(t)
	})

	// Test static method with invalid name - should panic
	t.Run("RegisterStatic invalid name", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic for invalid static method name, but no panic occurred")
			}
		}()

		testRegisterStaticInvalidName(t)
	})

	// Test static method with invalid function - should panic
	t.Run("RegisterStatic invalid function", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic for invalid static method function, but no panic occurred")
			}
		}()

		testRegisterStaticInvalidFunction(t)
	})
}

func TestRegisterMethod(t *testing.T) {
	// Register a method with documentation
	method := pobj.RegisterMethod("test/method-test:getInfo", func(ctx context.Context, id string) (string, error) {
		return "info for " + id, nil
	})

	if method == nil {
		t.Fatal("RegisterMethod returned nil")
	}

	// Verify the method is accessible
	obj := pobj.Get("test/method-test")
	if obj == nil {
		t.Fatal("Failed to get object with method")
	}

	// Get the method via Object.Method()
	retrieved := obj.Method("getInfo")
	if retrieved == nil {
		t.Fatal("Failed to get method via Object.Method()")
	}

	// Verify that Object.Static() also works (backward compatibility)
	callable := obj.Static("getInfo")
	if callable == nil {
		t.Fatal("Object.Static() should also return the callable")
	}

	// Call the method
	result, err := typutil.Call[string](callable, context.Background(), "test-id")
	if err != nil {
		t.Fatalf("Failed to call method: %v", err)
	}

	if result != "info for test-id" {
		t.Errorf("Wrong result, got %s, want %s", result, "info for test-id")
	}
}

func TestMethodDocumentation(t *testing.T) {
	// Register a method and set documentation
	method := pobj.RegisterMethod("test/method-doc:getUser", func(ctx context.Context) error {
		return nil
	}).SetDoc("Retrieves the current user from the system")

	if method == nil {
		t.Fatal("Method should not be nil after SetDoc")
	}

	// Verify the documentation is stored
	if method.Doc() != "Retrieves the current user from the system" {
		t.Errorf("Wrong doc, got %q, want %q", method.Doc(), "Retrieves the current user from the system")
	}

	// Verify it's accessible via the object
	obj := pobj.Get("test/method-doc")
	if obj == nil {
		t.Fatal("Failed to get object")
	}

	retrieved := obj.Method("getUser")
	if retrieved == nil {
		t.Fatal("Failed to get method")
	}

	if retrieved.Doc() != "Retrieves the current user from the system" {
		t.Errorf("Retrieved method has wrong doc, got %q", retrieved.Doc())
	}
}

func TestMethodRequiresInstance(t *testing.T) {
	// Register a method that requires an instance
	method := pobj.RegisterMethod("test/method-instance:save", func(ctx context.Context) error {
		return nil
	}).SetRequiresInstance(true).SetDoc("Saves the current object")

	if method == nil {
		t.Fatal("Method should not be nil")
	}

	if !method.RequiresInstance() {
		t.Error("Expected RequiresInstance to be true")
	}

	// Register another method that doesn't require instance
	staticMethod := pobj.RegisterMethod("test/method-instance:list", func(ctx context.Context) error {
		return nil
	})

	if staticMethod.RequiresInstance() {
		t.Error("Expected RequiresInstance to be false by default")
	}
}

func TestMethodAccessors(t *testing.T) {
	method := pobj.RegisterMethod("test/method-accessors:testMethod", func() {})

	// Test Name()
	if method.Name() != "testMethod" {
		t.Errorf("Wrong name, got %s, want %s", method.Name(), "testMethod")
	}

	// Test Object()
	obj := method.Object()
	if obj == nil {
		t.Fatal("Method.Object() returned nil")
	}

	if obj.String() != "test/method-accessors" {
		t.Errorf("Wrong object path, got %s, want %s", obj.String(), "test/method-accessors")
	}

	// Test String()
	if method.String() != "test/method-accessors:testMethod" {
		t.Errorf("Wrong string, got %s, want %s", method.String(), "test/method-accessors:testMethod")
	}

	// Test Callable()
	if method.Callable() == nil {
		t.Error("Callable() returned nil")
	}
}

func TestObjectMethods(t *testing.T) {
	// Register multiple methods on an object
	pobj.RegisterMethod("test/object-methods:method1", func() {})
	pobj.RegisterMethod("test/object-methods:method2", func() {})
	pobj.RegisterMethod("test/object-methods:method3", func() {})

	obj := pobj.Get("test/object-methods")
	if obj == nil {
		t.Fatal("Failed to get object")
	}

	methods := obj.Methods()
	if len(methods) != 3 {
		t.Errorf("Wrong number of methods, got %d, want %d", len(methods), 3)
	}

	// Check that all methods are present (order is not guaranteed)
	methodSet := make(map[string]bool)
	for _, name := range methods {
		methodSet[name] = true
	}

	for _, name := range []string{"method1", "method2", "method3"} {
		if !methodSet[name] {
			t.Errorf("Method %s not found in methods list", name)
		}
	}
}

func TestNilMethodSafety(t *testing.T) {
	var nilMethod *pobj.Method

	// All accessor methods should handle nil safely
	if nilMethod.Doc() != "" {
		t.Error("Doc() on nil should return empty string")
	}

	if nilMethod.Name() != "" {
		t.Error("Name() on nil should return empty string")
	}

	if nilMethod.String() != "" {
		t.Error("String() on nil should return empty string")
	}

	if nilMethod.RequiresInstance() != false {
		t.Error("RequiresInstance() on nil should return false")
	}

	if nilMethod.Callable() != nil {
		t.Error("Callable() on nil should return nil")
	}

	if nilMethod.Object() != nil {
		t.Error("Object() on nil should return nil")
	}

	// Setters should return nil safely
	if nilMethod.SetDoc("test") != nil {
		t.Error("SetDoc() on nil should return nil")
	}

	if nilMethod.SetRequiresInstance(true) != nil {
		t.Error("SetRequiresInstance() on nil should return nil")
	}
}
