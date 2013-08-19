package taskies

import (
	"fmt"
	"io"
	"reflect"
)

func pipeProvider(ps providerSet, data interface{}) (Task, error) {
	val := reflect.ValueOf(data)

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

		tasks[i] = task.task
	}

	return PipeTask(tasks...), nil
}

func PipeTask(tasks ...Task) Task {
	return func(env *Env, in io.Reader, out, err io.Writer) error {
		ch := make(chan error)
		l := len(tasks)

		for i, t := range tasks {
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
				err := t(env.Child(), in, out, err)

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
}
