package src

import (
	"reflect"
)

type ProxyTask struct {
	*baseTask
	task Task
	data interface{}
}

func (t *ProxyTask) Run(ctxt RunContext) error {
	val := reflect.ValueOf(t.data)

	if val.Kind() == reflect.Map {
		keys := val.MapKeys()

		for _, k := range keys {
			ks := k.String()
			vs := val.MapIndex(k).Elem().Interface()

			ctxt.Env().Set(ks, vs)
		}
	}

	Debugf("[PROXY TASK] %#v %v", t.task, t.data)

	return t.task.Run(ctxt)
}

func (t *ProxyTask) ExportData() []map[string]interface{} {
	return append(t.task.ExportData(), t.baseTask.ExportData()...)
}

func proxyProviderFunc(t Task) provider {
	return func(ps providerSet, data *taskData) (Task, error) {
		Debugf("[PROXY PROVIDER] %v", data)

		return &ProxyTask{
			baseTask: baseTaskFromTaskData(data),
			data:     data.taskData,
			task:     t,
		}, nil
	}
}
