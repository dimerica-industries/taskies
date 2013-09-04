package src

import (
	"testing"
)

func TestInheritance(t *testing.T) {
	e1 := NewEnv()
	e1.SetVar("a.b", "hello")

	e2 := e1.Child().Child()

	str := e2.GetVar("a.b").(string)

	if str != "hello" {
		t.Fatal()
	}
}
