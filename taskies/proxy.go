package taskies

import (
	"fmt"
	"reflect"
)

type ProxyTask struct {
	*baseTask
	task Task
	data interface{}
}

func (t *ProxyTask) Run(ctxt *RunContext) error {
	val := reflect.ValueOf(t.data)

	if val.Kind() == reflect.Map {
		keys := val.MapKeys()

		for _, k := range keys {
			ks := k.Elem().String()
			vs := fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface())

			ctxt.Env.Set(ks, vs)
		}
	}

	return ctxt.Run(t.task)
}

func proxyProviderFunc(t Task) provider {
	return func(ps providerSet, data *taskData) (Task, error) {
		return &ProxyTask{
			baseTask: baseTaskFromTaskData(data),
			data:     data.data,
			task:     t,
		}, nil
	}
}
