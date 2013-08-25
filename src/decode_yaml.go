package src

import (
	"fmt"
	"launchpad.net/goyaml"
	"reflect"
)

func DecodeYAML(contents []byte, ns *Namespace) error {
	var data interface{}
	err := goyaml.Unmarshal(contents, &data)

	data = clean(data)

	if err != nil {
		return err
	}

	val := reflect.ValueOf(data)
	d := newDecoder(ns)

	return d.decode(val)
}

func clean(val interface{}) interface{} {
	rv, ok := val.(reflect.Value)

	if !ok {
		rv = reflect.ValueOf(val)
	}

	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Map {
		m2 := make(map[string]interface{})
		keys := rv.MapKeys()

		for _, rk := range keys {
			ks := fmt.Sprintf("%v", rk.Interface())
			m2[ks] = clean(rv.MapIndex(rk))
		}

		return m2
	}

	if rv.Kind() == reflect.Slice {
		l := rv.Len()
		sl := make([]interface{}, l)

		for i := 0; i < l; i++ {
			sl[i] = clean(rv.Index(i))
		}

		return sl
	}

	return fmt.Sprintf("%v", rv.Interface())
}
