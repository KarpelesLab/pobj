package pobj_test

import (
	"context"
	"strings"
	"testing"

	"github.com/KarpelesLab/pobj"
	"github.com/KarpelesLab/typutil"
)

// TestEdgeCases focuses on testing the edge cases that weren't covered by the main tests
func TestEdgeCases(t *testing.T) {
	// Note: We can't easily test New() with nil type since the Object struct
	// is private and we can't set typ=nil from tests. This is covered by the
	// implementation tests.

	// Test Static() with nil static map
	t.Run("Static with nil static map", func(t *testing.T) {
		// Create a new object without static methods
		pobj.Register[struct{}]("edge_test/no_static")
		o := pobj.Get("edge_test/no_static")
		if o == nil {
			t.Fatal("Failed to get edge_test/no_static object")
		}

		// Get a static method - should return nil
		method := o.Static("anything")
		if method != nil {
			t.Errorf("Expected nil for Static() on object with no static methods, got %v", method)
		}
	})

	// Test ById with bad return type
	t.Run("ById with bad return type", func(t *testing.T) {
		// Register a type with an action that returns the wrong type
		actions := &pobj.ObjectActions{
			Fetch: typutil.Func(func(ctx context.Context, id string) (string, error) {
				// This returns a string, not a *BadType
				return "wrong type", nil
			}),
		}

		type BadType struct{}
		pobj.RegisterActions[BadType]("edge_test/bad_type", actions)

		// Try to fetch by ID - should fail with type error
		_, err := pobj.ById[BadType](context.Background(), "any-id")
		if err == nil {
			t.Error("Expected error for bad return type, got nil")
		}

		// Check that the error message contains expected information
		expectedErrSubstring := "bad type returned by Fetch"
		if err != nil && !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Error doesn't contain expected substring. Got: %v, want to contain: %s",
				err, expectedErrSubstring)
		}
	})
}
