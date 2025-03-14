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
