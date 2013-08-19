package taskies

import (
	"fmt"
	"io"
)

type Task interface {
	Name() string
	Description() string
	Run(*RunContext) error
}

type RunContext struct {
	Env *Env
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (c *RunContext) Run(t Task) error {
	ctxt := c.Clone()
	ctxt.Env = ctxt.Env.Child()
	Debugf("[task] [%s]", t.Name())

	return t.Run(ctxt)
}

func (c *RunContext) Clone() *RunContext {
	return &RunContext{
		Env: c.Env,
		In:  c.In,
		Out: c.Out,
		Err: c.Err,
	}
}

func NewRunner(tasks map[string]Task, env *Env, in io.Reader, out, err io.Writer) *Runner {
	return &Runner{
		tasks: tasks,
		context: &RunContext{
			Env: env,
			In:  in,
			Out: out,
			Err: err,
		},
	}
}

type Runner struct {
	tasks   map[string]Task
	context *RunContext
}

func (r *Runner) RunAll() error {
	for _, t := range r.tasks {
		if err := r.context.Run(t); err != nil {
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

		if err := r.context.Run(task); err != nil {
			return err
		}
	}

	return nil
}
