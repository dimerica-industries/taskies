package taskies

import (
	"fmt"
	"io"
	"reflect"
)

func RunMany(tasks []Task, env *Env, in io.Reader, out, err io.Writer) error {
	for _, t := range tasks {
		if e := t(env, in, out, err); e != nil {
			return e
		}
	}

	return nil
}

func CompositeProvider(ps ProviderSet, data interface{}) (Task, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("CompositeProvider expects slice, %s found", val.Kind())
	}

	tasks := make([]Task, val.Len())

	for i := 0; i < val.Len(); i++ {
		d := val.Index(i).Elem().Interface()
		task, err := ps.Provide(d)

		if err != nil {
			return nil, err
		}

		tasks[i] = task.task
	}

	return func(env *Env, in io.Reader, out, err io.Writer) error {
		return RunMany(tasks, env, in, out, err)
	}, nil
}
