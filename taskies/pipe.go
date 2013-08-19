package taskies

import (
	"fmt"
	"io"
	"reflect"
)

type PipeTask struct {
	name        string
	description string
	tasks       []Task
}

func (t *PipeTask) Name() string {
	return t.name
}

func (t *PipeTask) Description() string {
	return t.description
}

func (t *PipeTask) Run(env *Env, in io.Reader, out, err io.Writer) error {
	ch := make(chan error)
	l := len(t.tasks)

	for i, t := range t.tasks {
		var (
			pr io.Reader
			pw io.Writer
		)

		if i == l-1 {
			pr = in
			pw = out
		} else {
			pr, pw = io.Pipe()
		}

		go func(t Task, in io.Reader, out io.Writer) {
			err := run(t, env, in, out, err)

			if c, ok := out.(io.Closer); ok {
				c.Close()
			}

			ch <- err
		}(t, in, pw)

		in = pr
	}

	i := 0
	for err := range ch {
		if err != nil {
			return err
		}

		i++

		if i == l {
			return nil
		}
	}

	return nil
}

func pipeProvider(ps providerSet, data *taskData) (Task, error) {
	val := reflect.ValueOf(data.data)

	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("PipeProvider expects slice, %s found", val.Kind())
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

	return &PipeTask{
		name:        data.name,
		description: data.description,
		tasks:       tasks,
	}, nil
}
