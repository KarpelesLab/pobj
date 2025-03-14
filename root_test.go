package pobj_test

import (
	"testing"

	"github.com/KarpelesLab/pobj"
)

func TestRoot(t *testing.T) {
	// Get the root object
	root := pobj.Root()
	if root == nil {
		t.Fatal("Root() returned nil")
	}

	// Register some objects
	pobj.Register[struct{}]("root-test/object1")
	pobj.Register[struct{}]("root-test/object2")

	// Test that root has the expected children
	child := root.Child("root-test")
	if child == nil {
		t.Fatal("Failed to get child 'root-test' from root")
	}

	// Test that child has the expected children
	subchild1 := child.Child("object1")
	if subchild1 == nil {
		t.Error("Failed to get child 'object1' from 'root-test'")
	}

	subchild2 := child.Child("object2")
	if subchild2 == nil {
		t.Error("Failed to get child 'object2' from 'root-test'")
	}

	// Test string representation of root
	rootString := root.String()
	if rootString != "" {
		t.Errorf("Wrong string for root, got %q, want empty string", rootString)
	}
}

func TestErrors(t *testing.T) {
	// Verify error messages
	if pobj.ErrUnknownType.Error() != "pobj: unknown object type" {
		t.Errorf("Wrong error message for ErrUnknownType: %q", pobj.ErrUnknownType.Error())
	}

	if pobj.ErrMissingAction.Error() != "pobj: no such action exists" {
		t.Errorf("Wrong error message for ErrMissingAction: %q", pobj.ErrMissingAction.Error())
	}
}
