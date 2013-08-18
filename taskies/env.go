package taskies

import (
	"strings"
	"sync"
)

func FromArray(arr []string) *Env {
	env := NewEnv()

	for _, v := range arr {
		parts := strings.SplitN(v, "=", 2)
		key := parts[0]
		val := ""

		if len(parts) == 2 {
			val = parts[1]
		}

		env.vals[key] = val
	}

	return env
}

func NewEnv() *Env {
	return &Env{
		vals: make(map[string]string),
	}
}

type Env struct {
	l    sync.Mutex
	vals map[string]string
}

func (e *Env) Get(k string) string {
	e.l.Lock()
	defer e.l.Unlock()

	return e.vals[k]
}

func (e *Env) Set(k string, v string) {
	e.l.Lock()
	defer e.l.Unlock()

	e.vals[k] = v
}

func (v Env) Array() []string {
	a := make([]string, len(v.vals))
	i := 0

	for k, v := range v.vals {
		a[i] = k + "=" + v
		i++
	}

	return a
}

func MergeEnv(one Env, others ...Env) *Env {
	env := NewEnv()

	for k, v := range one.vals {
		env.vals[k] = v
	}

	for _, env2 := range others {
		for k, v := range env2.vals {
			env.vals[k] = v
		}
	}

	return env
}
