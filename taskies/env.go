package taskies

import (
	"sync"
)

func NewEnv() *Env {
	return &Env{
		vals: make(map[string]string),
	}
}

type Env struct {
	parent *Env
	l      sync.Mutex
	vals   map[string]string
}

func (e *Env) Get(k string) string {
	e.l.Lock()
	defer e.l.Unlock()

	v, ok := e.vals[k]

	if ok || e.IsRoot() {
		return v
	}

	return e.Parent().Get(k)
}

func (e *Env) Set(k string, v string) {
	e.l.Lock()
	defer e.l.Unlock()

	e.vals[k] = template(v, e)
}

func (e *Env) IsRoot() bool {
	return e.parent == nil
}

func (e *Env) Parent() *Env {
	return e.parent
}

func (e *Env) Root() *Env {
	if e.IsRoot() {
		return e
	}

	return e.Parent().Root()
}

func (e *Env) Child() *Env {
	e2 := NewEnv()
	e2.parent = e

	return e2
}
