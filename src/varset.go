package src

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func newVarSet() *varSet {
	return &varSet{
		vals: make(map[string]interface{}),
	}
}

// thread safe map of string -> interface{} with parent/child
// scope
//
// "." is used as an object delimeter, so get("a.b") is mapped to
// get("a").get("b")
type varSet struct {
	l        sync.RWMutex
	vals     map[string]interface{}
	exported map[string]bool
}

func (e *varSet) Get(k string) interface{} {
	e.l.RLock()
	defer e.l.RUnlock()

	return e.get(k)
}

func (e *varSet) get(k string) interface{} {
	if k == "." {
		return e.vals
	}

	var cur interface{} = e.vals
	parts := strings.Split(k, ".")

	for i, p := range parts {
		if e2, ok := cur.(*varSet); ok {
			return e2.Get(strings.Join(parts[i:], "."))
		}

		r := reflect.ValueOf(cur)
		kind := r.Kind()

		switch {
		case kind == reflect.Map:
			v := r.MapIndex(reflect.ValueOf(p))

			if !v.IsValid() {
				return nil
			}

			cur = v.Interface()
		case kind == reflect.Slice:
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

func (e *varSet) Set(k string, v interface{}) {
	e.l.Lock()
	defer e.l.Unlock()

	e.set(k, v)
}

func (e *varSet) set(k string, v interface{}) {
	rv := reflect.ValueOf(v)

	parts := strings.Split(k, ".")
	l := len(parts)

	if l == 0 {
		return
	}

	cur := reflect.ValueOf(e.vals)

	for i, p := range parts {
		if e2, ok := cur.Interface().(*varSet); ok {
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
