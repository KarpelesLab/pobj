package pobj_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/KarpelesLab/pobj"
	"github.com/KarpelesLab/typutil"
)

// TestPerson is used for testing object registration and retrieval
type TestPerson struct {
	ID    string
	Name  string
	Email string
}

// TestCompany is another test type
type TestCompany struct {
	ID      string
	Name    string
	Address string
}

// We need to create unique registration paths for each test to avoid
// the "multiple registrations" panic
var testPathCounter = 0

// Setup function for registering test objects
func setupTestObjects() string {
	// Generate a unique prefix for this test run
	testPathCounter++
	prefix := "test" + string(rune(testPathCounter+'0')) // Convert to a rune

	personPath := prefix + "/person"
	companyPath := prefix + "/company"

	// Register a person type
	pobj.Register[TestPerson](personPath)

	// Register a company type with actions
	actions := &pobj.ObjectActions{
		Fetch: typutil.Func(func(ctx context.Context, id string) (*TestCompany, error) {
			return &TestCompany{ID: id, Name: "Test Company", Address: "123 Test St"}, nil
		}),
		List: typutil.Func(func(ctx context.Context) ([]*TestCompany, error) {
			return []*TestCompany{
				{ID: "1", Name: "Company A", Address: "Address A"},
				{ID: "2", Name: "Company B", Address: "Address B"},
			}, nil
		}),
		Create: typutil.Func(func(ctx context.Context, data *TestCompany) (*TestCompany, error) {
			return data, nil
		}),
		Clear: typutil.Func(func(ctx context.Context) error {
			return nil
		}),
	}
	pobj.RegisterActions[TestCompany](companyPath, actions)

	// Register static methods
	pobj.RegisterStatic(personPath+":getByEmail", func(ctx context.Context, email string) (*TestPerson, error) {
		return &TestPerson{ID: "test-id", Name: "Test User", Email: email}, nil
	})

	return prefix
}

func TestRegisterAndGet(t *testing.T) {
	prefix := setupTestObjects()

	tests := []struct {
		name       string
		objectPath string
		wantType   reflect.Type
		wantNil    bool
	}{
		{
			name:       "Get registered person",
			objectPath: prefix + "/person",
			wantType:   reflect.TypeOf(TestPerson{}),
			wantNil:    false,
		},
		{
			name:       "Get registered company",
			objectPath: prefix + "/company",
			wantType:   reflect.TypeOf(TestCompany{}),
			wantNil:    false,
		},
		{
			name:       "Get non-existent object",
			objectPath: prefix + "/nonexistent",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pobj.Get(tt.objectPath)

			if tt.wantNil {
				if got != nil {
					t.Errorf("Expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("Expected non-nil, got nil")
			}

			// Create a new instance to check the type
			instance := got.New()
			if instance == nil {
				t.Fatalf("Failed to create new instance")
			}

			// Get the type of the value that the interface{} contains
			instanceType := reflect.TypeOf(instance).Elem()
			if instanceType != tt.wantType {
				t.Errorf("Wrong type, got %v, want %v", instanceType, tt.wantType)
			}
		})
	}
}

func TestGetByType(t *testing.T) {
	prefix := setupTestObjects()
	personPath := prefix + "/person"

	t.Run("Get existing type", func(t *testing.T) {
		obj := pobj.GetByType[TestPerson]()
		if obj == nil {
			t.Fatal("Expected to get object for TestPerson, got nil")
		}

		if obj.String() != personPath {
			t.Errorf("Wrong object path, got %s, want %s", obj.String(), personPath)
		}
	})

	t.Run("Get non-existent type", func(t *testing.T) {
		type UnregisteredType struct{}
		obj := pobj.GetByType[UnregisteredType]()
		if obj != nil {
			t.Errorf("Expected nil for unregistered type, got %v", obj)
		}
	})
}

func TestObjectNew(t *testing.T) {
	prefix := setupTestObjects()
	personPath := prefix + "/person"

	obj := pobj.Get(personPath)
	if obj == nil {
		t.Fatalf("Failed to get object at path: %s", personPath)
	}

	instance := obj.New()
	if instance == nil {
		t.Fatal("Failed to create new instance")
	}

	_, ok := instance.(*TestPerson)
	if !ok {
		t.Errorf("Wrong type, expected *TestPerson, got %T", instance)
	}
}

func TestObjectString(t *testing.T) {
	// For this test, we'll create our own objects to avoid conflicts
	testPrefix := "string_test"

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Root level object",
			path:     testPrefix,
			expected: testPrefix,
		},
		{
			name:     "Child object",
			path:     testPrefix + "/child",
			expected: testPrefix + "/child",
		},
		{
			name:     "Nested object",
			path:     testPrefix + "/sub/nested",
			expected: testPrefix + "/sub/nested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the object to ensure it exists
			pobj.Register[struct{}](tt.path)

			obj := pobj.Get(tt.path)
			if obj == nil {
				t.Fatalf("Failed to get object at path: %s", tt.path)
			}

			result := obj.String()
			if result != tt.expected {
				t.Errorf("Wrong string representation, got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestChild(t *testing.T) {
	prefix := setupTestObjects()

	root := pobj.Get(prefix)
	if root == nil {
		t.Fatalf("Failed to get root object at path: %s", prefix)
	}

	t.Run("Get existing child", func(t *testing.T) {
		child := root.Child("person")
		if child == nil {
			t.Error("Expected to get child 'person', got nil")
		}
	})

	t.Run("Get non-existent child", func(t *testing.T) {
		child := root.Child("nonexistent")
		if child != nil {
			t.Errorf("Expected nil for non-existent child, got %v", child)
		}
	})

	t.Run("Child of nil", func(t *testing.T) {
		var nilObj *pobj.Object
		child := nilObj.Child("anything")
		if child != nil {
			t.Errorf("Expected nil when calling Child on nil Object, got %v", child)
		}
	})
}

func TestObjectDocumentation(t *testing.T) {
	// Register an object with documentation
	obj := pobj.Register[struct{}]("test/object-doc").
		SetDoc("This object represents a test entity for documentation purposes")

	if obj == nil {
		t.Fatal("Register returned nil")
	}

	// Verify the documentation is stored
	if obj.Doc() != "This object represents a test entity for documentation purposes" {
		t.Errorf("Wrong doc, got %q", obj.Doc())
	}

	// Verify it's retrievable via Get
	retrieved := pobj.Get("test/object-doc")
	if retrieved == nil {
		t.Fatal("Failed to get object")
	}

	if retrieved.Doc() != "This object represents a test entity for documentation purposes" {
		t.Errorf("Retrieved object has wrong doc, got %q", retrieved.Doc())
	}
}

func TestNilObjectSafety(t *testing.T) {
	var nilObj *pobj.Object

	// Doc methods should handle nil safely
	if nilObj.Doc() != "" {
		t.Error("Doc() on nil should return empty string")
	}

	if nilObj.SetDoc("test") != nil {
		t.Error("SetDoc() on nil should return nil")
	}

	// Method and Methods should handle nil safely
	if nilObj.Method("anything") != nil {
		t.Error("Method() on nil should return nil")
	}

	if nilObj.Methods() != nil {
		t.Error("Methods() on nil should return nil")
	}
}
