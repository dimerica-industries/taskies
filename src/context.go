package src

import (
	"bytes"
	"io"
)

type baseContext struct {
	env        *Env
	in         io.Reader
	out        io.Writer
	err        io.Writer
	childTasks []Task
}

func (c *baseContext) Env() *Env {
	return c.env
}

func (c *baseContext) In() io.Reader {
	return c.in
}

func (c *baseContext) Out() io.Writer {
	return c.out
}

func (c *baseContext) Err() io.Writer {
	return c.err
}

func (c *baseContext) Run(t Task) *RunResult {
	out := new(bytes.Buffer)
	er := new(bytes.Buffer)

	c.childTasks = append(c.childTasks, t)

	env := c.Env().Child()

	c.Env().Set(t.Name(), env)

	sout := io.MultiWriter(c.Out(), out)
	serr := io.MultiWriter(c.Err(), er)

	ctxt := c.Clone(env, nil, sout, serr)

	Debugf("[RUN] [name=%s] [env=%s]", t.Name(), ctxt.Env().Id())

	err := t.Run(ctxt)

	ctxt.Env().Set("$stdout", string(out.Bytes()))
	ctxt.Env().Set("$stderr", string(er.Bytes()))

	if err != nil {
		ctxt.Env().Set("$error", err.Error())
	}

	exports := t.ExportData()

	for _, export := range exports {
		for k, v := range export {
			ctxt.Env().Set(k, v)
		}
	}

	c.Env().Set("$last", env)

	return &RunResult{
		out:   out.Bytes(),
		err:   er.Bytes(),
		error: err,
	}
}

func (c *baseContext) Clone(env *Env, in io.Reader, out, err io.Writer) RunContext {
	if env == nil {
		env = c.env
	}

	if in == nil {
		in = c.in
	}

	if out == nil {
		out = c.out
	}

	if err == nil {
		err = c.err
	}

	return &baseContext{
		env:        env,
		in:         in,
		out:        out,
		err:        err,
		childTasks: make([]Task, 0),
	}
}
