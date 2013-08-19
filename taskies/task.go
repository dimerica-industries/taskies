package taskies

import (
	"bytes"
	"fmt"
	"io"
)

type Task interface {
	Name() string
	Description() string
	Run(*RunContext) error
	EnvSet() map[string]string
}

type baseTask struct {
	name        string
	description string
	envSet      map[string]string
}

func (t *baseTask) Name() string {
	return t.name
}

func (t *baseTask) Description() string {
	return t.description
}

func (t *baseTask) EnvSet() map[string]string {
	return t.envSet
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

	out := new(bytes.Buffer)
	er := new(bytes.Buffer)

	ctxt.Out = io.MultiWriter(ctxt.Out, out)
	ctxt.Err = io.MultiWriter(ctxt.Err, er)

	err := t.Run(ctxt)

	c.Env.Set("$result.stdout", string(out.Bytes()))
	c.Env.Set("$result.stderr", string(er.Bytes()))

	if err != nil {
		c.Env.Set("$result.error", err.Error())
	}

	set := t.EnvSet()

	if set != nil {
		for k, v := range set {
			c.Env.Set(k, v)
		}
	}

	return err
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
