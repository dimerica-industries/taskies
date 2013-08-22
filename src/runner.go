package src

import (
	"fmt"
	"io"
	"strings"
)

func NewRunner(tasks *TaskSet, env *Env, in io.Reader, out, err io.Writer) *Runner {
	return &Runner{
		tasks:   tasks,
		context: newBaseContext(env, in, out, err),
	}
}

type Runner struct {
	tasks   *TaskSet
	context *baseContext
}

func (r *Runner) RunAll() error {
	for _, t := range r.tasks.ExportedTasks {
		if res := r.context.Run(t); res.error != nil {
			return res.error
		}
	}

	return nil
}

func (r *Runner) Run(tasks ...string) error {
	for _, t := range tasks {
		t := strings.ToLower(t)
		task, ok := r.tasks.ExportedTasks[t]

		if !ok {
			return fmt.Errorf("Missing task %s", t)
		}

		if res := r.context.Run(task); res.error != nil {
			return res.error
		}
	}

	return nil
}
