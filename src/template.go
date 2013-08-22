package src

import (
	"fmt"
	"github.com/dimerica-industries/taskies/mustache"
	"reflect"
)

func template(tmpl interface{}, env *Env) interface{} {
	if _, ok := tmpl.(*Env); ok {
		return tmpl
	}

	var str string
	r := reflect.ValueOf(tmpl)

	switch r.Kind() {
	case reflect.Interface:
		return template(r.Elem().Interface(), env)
	case reflect.Map:
		m := make(map[string]interface{})
		keys := r.MapKeys()

		for _, k := range keys {
			v := r.MapIndex(k)
			tk := template(k, env).(string)
			tv := template(v.Interface(), env)

			m[tk] = tv
		}

		return m
	case reflect.Slice:
		l := r.Len()
		sl := make([]interface{}, l)

		for i := 0; i < l; i++ {
			v := r.Index(i).Elem().Interface()
			sl[i] = template(v, env)
		}

		return sl
	default:
		str = fmt.Sprintf("%v", tmpl)
	}

	out := mustache.Render(str, &finder{env})

	Debugf("[TEMPLATE] [env=%s] [before=%s] [after=%s]", env.Id(), str, out)

	return out
}

type finder struct {
	env *Env
}

func (f *finder) Lookup(name string) reflect.Value {
	if v := f.env.Get(name); v != nil {
		return reflect.ValueOf(v)
	}

	return reflect.Value{}
}
