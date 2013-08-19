package taskies

import (
	"fmt"
	"io"
	"reflect"
)

func proxyProviderFunc(t Task) provider {
	return func(ps providerSet, data interface{}) (Task, error) {
		return func(env *Env, in io.Reader, out, err io.Writer) error {
			val := reflect.ValueOf(data)

			if val.Kind() == reflect.Map {
				keys := val.MapKeys()

				for _, k := range keys {
					ks := k.Elem().String()
					vs := fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface())

					env.Set(ks, vs)
				}
			}

			return t(env, in, out, err)
		}, nil
	}
}
