package src

import (
	"fmt"
	"github.com/dimerica-industries/taskies/mustache"
	"reflect"
)

func template(tmpl interface{}, env *Env) interface{} {
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
			v := r.Index(i)
			sl[i] = template(v, env)
		}

		return sl
	default:
		str = fmt.Sprintf("%v", tmpl)
	}

	ctxt := []interface{}{env.vals}

	for !env.IsRoot() {
		env = env.Parent()
		ctxt = append(ctxt, env.vals)
	}

	Debugf("[TEMPLATE] %s %v", str, ctxt)

	return mustache.Render(str, ctxt...)
}
