package pobj_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/pobj"
	"github.com/KarpelesLab/typutil"
)

func TestStatic(t *testing.T) {
	prefix := setupTestObjects()
	personPath := prefix + "/person"

	t.Run("Get existing static method", func(t *testing.T) {
		obj := pobj.Get(personPath)
		if obj == nil {
			t.Fatalf("Failed to get object at path: %s", personPath)
		}

		method := obj.Static("getByEmail")
		if method == nil {
			t.Fatal("Expected to get 'getByEmail' static method, got nil")
		}

		// Test calling the static method
		result, err := typutil.Call[*TestPerson](method, context.Background(), "test@example.com")
		if err != nil {
			t.Fatalf("Failed to call static method: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result from static method")
		}

		if result.Email != "test@example.com" {
			t.Errorf("Wrong email in result, got %s, want %s", result.Email, "test@example.com")
		}
	})

	t.Run("Get non-existent static method", func(t *testing.T) {
		obj := pobj.Get(personPath)
		if obj == nil {
			t.Fatalf("Failed to get object at path: %s", personPath)
		}

		method := obj.Static("nonexistent")
		if method != nil {
			t.Errorf("Expected nil for non-existent static method, got %v", method)
		}
	})

	t.Run("Static on nil object", func(t *testing.T) {
		var nilObj *pobj.Object
		method := nilObj.Static("anything")
		if method != nil {
			t.Errorf("Expected nil when calling Static on nil Object, got %v", method)
		}
	})
}

// Test the deprecated static functions
func TestDeprecatedStatic(t *testing.T) {
	// Test Static function
	method := pobj.Static(func(ctx context.Context, arg string) (string, error) {
		return "Hello, " + arg, nil
	})

	if method == nil {
		t.Fatal("Expected non-nil result from Static")
	}

	// Test Call function
	result, err := pobj.Call[string](method, context.Background(), "World")
	if err != nil {
		t.Fatalf("Failed to call static method: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("Wrong result, got %s, want %s", result, "Hello, World")
	}
}
