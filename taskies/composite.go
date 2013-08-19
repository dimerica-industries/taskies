package taskies

import (
	"fmt"
	"io"
	"reflect"
)

func CompositeTask(tasks ...Task) Task {
	return func(env *Env, in io.Reader, out, err io.Writer) error {
		for _, t := range tasks {
			if e := t(env.Child(), in, out, err); e != nil {
				return e
			}
		}

		return nil
	}
}

func compositeProvider(ps providerSet, data interface{}) (Task, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("CompositeProvider expects slice, %s found", val.Kind())
	}

	tasks := make([]Task, val.Len())

	for i := 0; i < val.Len(); i++ {
		d := val.Index(i).Elem().Interface()
		task, err := ps.provide(d)

		if err != nil {
			return nil, err
		}

		tasks[i] = task.task
	}

	return CompositeTask(tasks...), nil
}
