package src

import (
	"fmt"
	"reflect"
	"testing"
)

var _ = fmt.Println

func TestEnvBasicSet(t *testing.T) {
	e := NewEnv()
	e.Set("a", "b")

	if e.Get("a") != "b" {
		t.Errorf("expected \"b\", found %s", e.Get("a"))
	}
}

func TestEnvNestedSet(t *testing.T) {
	e := NewEnv()
	e.Set("a.b", "c")

	if e.Get("a.b") != "c" {
		t.Fatalf("expected \"c\", found %s", e.Get("a.b"))
	}

	e.Set("a.b.c", "d")

	r := reflect.ValueOf(e.Get("a.b"))

	if r.Kind() != reflect.Map {
		t.Fatalf("expected type map, found %s", r.Kind())
	}

	e.Set("a.b.e", "f")

	if e.Get("a.b.c") != "d" {
		t.Fatalf("expected \"d\", found %s", e.Get("a.b.c"))
	}

	if e.Get("a.b.e") != "f" {
		t.Fatalf("expected \"f\", found %s", e.Get("a.b.e"))
	}
}

func TestChainedEnv(t *testing.T) {
	e1 := NewEnv()
	e2 := NewEnv()

	e2.Set("a.b.c", 3)
	e1.Set("d.e", e2)

	if e1.Get("d.e.a.b.c") != "3" {
		t.Fatalf("Expected 3,found %s", e1.Get("d.e.a.b.c"))
	}
}
