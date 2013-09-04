package src

import (
	"fmt"
	"reflect"
	"testing"
)

var _ = fmt.Println

func TestVarBasicSet(t *testing.T) {
	e := newVarSet()
	e.Set("a", "b")

	if e.Get("a") != "b" {
		t.Errorf("expected \"b\", found %s", e.Get("a"))
	}
}

func TestVarNestedSet(t *testing.T) {
	e := newVarSet()
	e.Set("a.b", "c")

	if e.Get("a.b") != "c" {
		t.Fatalf("expected \"c\", found %s", e.Get("a.b"))
	}

	e.Set("a.b.c", "d")

	r := reflect.ValueOf(e.Get("a.b"))

    if r.Kind() != reflect.Map {
		t.Fatalf("expected type VarSet, found %s", r.Kind())
	}

	e.Set("a.b.e", "f")

	if e.Get("a.b.c") != "d" {
		t.Fatalf("expected \"d\", found %s", e.Get("a.b.c"))
	}

	if e.Get("a.b.e") != "f" {
		t.Fatalf("expected \"f\", found %s", e.Get("a.b.e"))
	}
}
