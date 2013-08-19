package taskies

import (
	"fmt"
	"io"
	"reflect"
)

type CompositeTask struct {
	name        string
	description string
	tasks       []Task
}

func (t *CompositeTask) Name() string {
	return t.name
}

func (t *CompositeTask) Description() string {
	return t.description
}

func (t *CompositeTask) Run(env *Env, in io.Reader, out, err io.Writer) error {
	for _, t := range t.tasks {
		if e := run(t, env, in, out, err); e != nil {
			return e
		}
	}

	return nil
}

func compositeProvider(ps providerSet, data *taskData) (Task, error) {
	val := reflect.ValueOf(data.data)

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

		tasks[i] = task
	}

	return &CompositeTask{
		name:        data.name,
		description: data.description,
		tasks:       tasks,
	}, nil
}
