package handler_test

import (
	"testing"

	handler "github.com/holdemlab/ui-json-schema/api"
)

type testUser struct {
	Name  string `json:"name" required:"true"`
	Email string `json:"email" format:"email"`
	Age   int    `json:"age"`
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	r := handler.NewRegistry()
	r.Register("User", testUser{})

	v, err := r.Lookup("User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := v.(testUser); !ok {
		t.Errorf("expected testUser, got %T", v)
	}
}

func TestRegistry_LookupNotFound(t *testing.T) {
	r := handler.NewRegistry()

	_, err := r.Lookup("NonExistent")
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestRegistry_Overwrite(t *testing.T) {
	r := handler.NewRegistry()

	type v1 struct{ A string }
	type v2 struct{ B int }

	r.Register("Obj", v1{})
	r.Register("Obj", v2{})

	v, err := r.Lookup("Obj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := v.(v2); !ok {
		t.Errorf("expected v2 after overwrite, got %T", v)
	}
}

func TestRegistry_Names(t *testing.T) {
	r := handler.NewRegistry()
	r.Register("Beta", struct{}{})
	r.Register("Alpha", struct{}{})

	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	if !nameSet["Alpha"] || !nameSet["Beta"] {
		t.Errorf("expected Alpha and Beta in names, got %v", names)
	}
}

func TestRegistry_NamesEmpty(t *testing.T) {
	r := handler.NewRegistry()
	names := r.Names()

	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}
