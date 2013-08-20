package src

import (
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
	l      sync.Mutex
	vals   map[string]interface{}
}

func (e *Env) Get(k string) interface{} {
	e.l.Lock()
	defer e.l.Unlock()

    return e.get(k)
}

func (e *Env) get(k string) interface{} {
    var cur interface{} = e.vals
    parts := strings.Split(k, ".")

    for _, p := range parts {
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
	e.l.Lock()
	defer e.l.Unlock()

    e.set(k, v)
}

func (e *Env) set(k string, v interface{}) {
    k = template(k, e).(string)
    v = template(v, e)
    rv := reflect.ValueOf(v)

    parts := strings.Split(k, ".")
    l := len(parts)
    cur := reflect.ValueOf(e.vals)

    for i, p := range parts {
        rp := reflect.ValueOf(p)

        if i == l - 1 {
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
