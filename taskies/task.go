package taskies

import (
	"fmt"
	"io"
)

type Task func(env *Env, in io.Reader, out, err io.Writer) error

func NewRunner(tasks map[string]Task, env *Env, in io.Reader, out, err io.Writer) *Runner {
	return &Runner{
		tasks: tasks,
		env:   env,
		in:    in,
		out:   out,
		err:   err,
	}
}

type Runner struct {
	tasks map[string]Task
	env   *Env
	in    io.Reader
	out   io.Writer
	err   io.Writer
}

func (r *Runner) RunAll() error {
	for _, t := range r.tasks {
		env := r.env.Child()

		if err := t(env, r.in, r.out, r.err); err != nil {
			return err
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

		env := r.env.Child()

		if err := task(env, r.in, r.out, r.err); err != nil {
			return err
		}
	}

	return nil
}
