package src

import (
	"testing"
)

func TestInheritance(t *testing.T) {
	e1 := NewEnv()
	e1.SetVar("a.b.c.d", "hello")

	e2 := e1.Child().Child().Child()

	str := e2.GetVar("a.b.c.d").(string)

	if str != "hello" {
		t.Fatal()
	}
}
