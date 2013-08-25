package src

import (
	"fmt"
	"io"
	"reflect"
)

// Start multiple tasks in parallel but join stdout -> stdin of
// contiguous tasks
//
// Represented as "- pipe" in YAML
type PipeTask struct {
	*baseTask
	tasks []Task
}

func (t *PipeTask) Run(ctxt RunContext) error {
	ch := make(chan error)
	l := len(t.tasks)
	in := ctxt.In()

	for i, t := range t.tasks {
		var (
			pr io.Reader
			pw io.Writer
		)

		if i == l-1 {
			pr = ctxt.In()
			pw = ctxt.Out()
		} else {
			pr, pw = io.Pipe()
		}

		go func(t Task, in io.Reader, out io.Writer) {
			ctxt = ctxt.Clone(nil, in, out, nil)
			res := ctxt.Run(t)

			if c, ok := out.(io.Closer); ok {
				c.Close()
			}

			ch <- res.error
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

func pipeDecoder(ts taskDecoderSet) taskDecoder {
	return func(data *taskData) (Task, error) {
		val := reflect.ValueOf(data.taskData)

		if val.Kind() != reflect.Slice {
			return nil, fmt.Errorf("PipeDecoder expects slice, %s found", val.Kind())
		}

		tasks := make([]Task, val.Len())

		for i := 0; i < val.Len(); i++ {
			d := val.Index(i).Elem().Interface()
			task, err := ts.decode(d)

			if err != nil {
				return nil, err
			}

			tasks[i] = task
		}

		return &PipeTask{
			baseTask: baseTaskFromTaskData(data),
			tasks:    tasks,
		}, nil
	}
}
