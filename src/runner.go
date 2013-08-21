package src

import (
	"fmt"
	"io"
)

func NewRunner(tasks map[string]Task, env *Env, in io.Reader, out, err io.Writer) *Runner {
	return &Runner{
		tasks: tasks,
		context: &baseContext{
			env:        env,
			in:         in,
			out:        out,
			err:        err,
			childTasks: make([]Task, 0),
		},
	}
}

type Runner struct {
	tasks   map[string]Task
	context *baseContext
}

func (r *Runner) RunAll() error {
	for _, t := range r.tasks {
		if res := r.context.Run(t); res.error != nil {
			return res.error
		}
	}

	return nil
}

func (r *Runner) Run(tasks ...string) error {
	for _, t := range tasks {
		task, ok := r.tasks[t]

		if !ok {
			return fmt.Errorf("Missing task %s", t)
		}

		if res := r.context.Run(task); res.error != nil {
			return res.error
		}
	}

	return nil
}
