package src

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func NewEnv() *Env {
	return &Env{
		vals: make(map[string]interface{}),
	}
}

type Env struct {
	parent *Env
	l      sync.RWMutex
	vals   map[string]interface{}
}

func (e *Env) Id() string {
	id := fmt.Sprintf("%p", e)

	if e.IsRoot() {
		return id
	}

	return e.Parent().Id() + "." + id
}

func (e *Env) Get(k string) interface{} {
	e.l.RLock()
	defer e.l.RUnlock()

	v := e.get(k)

	if v != nil || e.IsRoot() {
		return v
	}

	return e.Parent().Get(k)
}

func (e *Env) get(k string) interface{} {
    if k == "." {
        return e.vals
    }

	var cur interface{} = e.vals
	parts := strings.Split(k, ".")

	for i, p := range parts {
		if e2, ok := cur.(*Env); ok {
			return e2.Get(strings.Join(parts[i:], "."))
		}

		r := reflect.ValueOf(cur)

		switch r.Kind() {
		case reflect.Map:
			v := r.MapIndex(reflect.ValueOf(p))

			if !v.IsValid() {
				return nil
			}

			cur = v.Interface()
		case reflect.Slice:
			i, _ := strconv.Atoi(p)
			v := r.Index(i)

			if !v.IsValid() {
				return nil
			}

			cur = v.Interface()
		default:
			return nil
		}
	}

	return cur

}

func (e *Env) Set(k string, v interface{}) {
	k = template(k, e).(string)
	v = template(v, e)

	e.l.Lock()
	defer e.l.Unlock()

	e.set(k, v)
}

func (e *Env) set(k string, v interface{}) {
	Debugf("[ENV SET] %s %#v = %#v", e.Id(), k, v)
	rv := reflect.ValueOf(v)

	parts := strings.Split(k, ".")
	l := len(parts)
	cur := reflect.ValueOf(e.vals)

	for i, p := range parts {
		if e2, ok := cur.Interface().(*Env); ok {
			e2.Set(strings.Join(parts[i:], "."), v)
			return
		}

		rp := reflect.ValueOf(p)

		if i == l-1 {
			cur.SetMapIndex(rp, rv)
			return
		}

		v := cur.MapIndex(rp)

		if v.IsValid() {
			v = v.Elem()
		}

		if !v.IsValid() || v.Kind() != reflect.Map {
			curv := make(map[string]interface{})
			tmp := reflect.ValueOf(curv)
			cur.SetMapIndex(rp, tmp)

			cur = tmp
		} else {
			cur = v
		}
	}
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

	Debugf("[NEW ENV] %s %s", e.Id(), e2.Id())

	return e2
}
