package src

import (
	"fmt"
	"github.com/dimerica-industries/taskies/mustache"
	"reflect"
)

func template(tmpl interface{}, e *Env) interface{} {
	if _, ok := tmpl.(*varSet); ok {
		return tmpl
	}

	var str string
	r := reflect.ValueOf(tmpl)

	switch r.Kind() {
	case reflect.Interface:
		return template(r.Elem().Interface(), e)
	case reflect.Map:
		m := make(map[string]interface{})
		keys := r.MapKeys()

		for _, k := range keys {
			v := r.MapIndex(k)
			tk := template(k, e).(string)
			tv := template(v.Interface(), e)

			m[tk] = tv
		}

		return m
	case reflect.Slice:
		l := r.Len()
		sl := make([]interface{}, l)

		for i := 0; i < l; i++ {
			v := r.Index(i).Elem().Interface()
			sl[i] = template(v, e)
		}

		return sl
	default:
		str = fmt.Sprintf("%v", tmpl)
	}

	out := mustache.Render(str, &finder{e})

	Debugf("[TEMPLATE] [ENV=%s] [before=%s] [after=%s]", e.Id(), str, out)

	return out
}

type finder struct {
	env *Env
}

func (f *finder) Lookup(name string) reflect.Value {
	if v := f.env.GetVar(name); v != nil {
		return reflect.ValueOf(v)
	}

	return reflect.Value{}
}
