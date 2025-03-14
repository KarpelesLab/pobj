package pobj_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/pobj"
)

func TestById(t *testing.T) {
	_ = setupTestObjects() // We don't need the prefix for this test
	ctx := context.Background()

	t.Run("Fetch existing object by ID", func(t *testing.T) {
		company, err := pobj.ById[TestCompany](ctx, "test-id-123")
		if err != nil {
			t.Fatalf("Failed to fetch company by ID: %v", err)
		}

		if company == nil {
			t.Fatal("Expected non-nil company")
		}

		if company.ID != "test-id-123" {
			t.Errorf("Wrong ID, got %s, want %s", company.ID, "test-id-123")
		}

		if company.Name != "Test Company" {
			t.Errorf("Wrong name, got %s, want %s", company.Name, "Test Company")
		}
	})

	t.Run("Fetch object by ID with unknown type", func(t *testing.T) {
		type UnknownType struct{}

		_, err := pobj.ById[UnknownType](ctx, "any-id")
		if err == nil {
			t.Error("Expected error for unknown type, got nil")
		}

		if err != pobj.ErrUnknownType {
			t.Errorf("Wrong error, got %v, want %v", err, pobj.ErrUnknownType)
		}
	})
}

func TestObjectById(t *testing.T) {
	prefix := setupTestObjects()
	ctx := context.Background()
	companyPath := prefix + "/company"
	personPath := prefix + "/person"

	t.Run("Fetch using Object.ById", func(t *testing.T) {
		obj := pobj.Get(companyPath)
		if obj == nil {
			t.Fatalf("Failed to get object at path: %s", companyPath)
		}

		result, err := obj.ById(ctx, "test-id-456")
		if err != nil {
			t.Fatalf("Failed to fetch by ID: %v", err)
		}

		company, ok := result.(*TestCompany)
		if !ok {
			t.Fatalf("Wrong type, got %T, want *TestCompany", result)
		}

		if company.ID != "test-id-456" {
			t.Errorf("Wrong ID, got %s, want %s", company.ID, "test-id-456")
		}
	})

	t.Run("Fetch with missing action", func(t *testing.T) {
		// Person doesn't have actions
		obj := pobj.Get(personPath)
		if obj == nil {
			t.Fatalf("Failed to get object at path: %s", personPath)
		}

		_, err := obj.ById(ctx, "any-id")
		if err == nil {
			t.Error("Expected error for missing action, got nil")
		}

		if err != pobj.ErrMissingAction {
			t.Errorf("Wrong error, got %v, want %v", err, pobj.ErrMissingAction)
		}
	})

	t.Run("Fetch with nil Fetch action", func(t *testing.T) {
		// Register an object with nil Fetch action
		noFetchPath := prefix + "/no-fetch"
		actions := &pobj.ObjectActions{
			List:   nil,
			Create: nil,
			Clear:  nil,
			// Fetch is nil
		}
		pobj.RegisterActions[struct{}](noFetchPath, actions)

		obj := pobj.Get(noFetchPath)
		if obj == nil {
			t.Fatalf("Failed to get object at path: %s", noFetchPath)
		}

		_, err := obj.ById(ctx, "any-id")
		if err == nil {
			t.Error("Expected error for nil Fetch action, got nil")
		}

		if err != pobj.ErrMissingAction {
			t.Errorf("Wrong error, got %v, want %v", err, pobj.ErrMissingAction)
		}
	})
}
