package taskies

import (
	"fmt"
	"io"
	"reflect"
)

type ProxyTask struct {
	name        string
	description string
	task        Task
	data        interface{}
}

func (t *ProxyTask) Name() string {
	return t.name
}

func (t *ProxyTask) Description() string {
	return t.description
}

func (t *ProxyTask) Run(env *Env, in io.Reader, out, err io.Writer) error {
	val := reflect.ValueOf(t.data)

	if val.Kind() == reflect.Map {
		keys := val.MapKeys()

		for _, k := range keys {
			ks := k.Elem().String()
			vs := fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface())

			env.Set(ks, vs)
		}
	}

	return run(t.task, env, in, out, err)
}

func proxyProviderFunc(t Task) provider {
	return func(ps providerSet, data *taskData) (Task, error) {
		return &ProxyTask{
			name:        data.name,
			description: data.description,
			data:        data.data,
			task:        t,
		}, nil
	}
}
