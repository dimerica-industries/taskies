package src

import (
	"io"
)

type Task interface {
	Name() string
	Type() string
	Description() string
	Run(RunContext) error
	Export() map[string]interface{}
	Var() string
}

type RunContext interface {
	In() io.Reader
	Out() io.Writer
	Err() io.Writer
	Run(Task) error
	Clone(io.Reader, io.Writer, io.Writer) RunContext
	Env() *Env
}

type context struct {
	in    io.Reader
	out   io.Writer
	err   io.Writer
	runfn func(RunContext, Task) error
	env   *Env
}

func (c *context) Env() *Env {
	return c.env
}

func (c *context) In() io.Reader {
	return c.in
}

func (c *context) Out() io.Writer {
	return c.out
}

func (c *context) Err() io.Writer {
	return c.err
}

func (c *context) Run(t Task) error {
	return c.runfn(c, t)
}

func (c *context) Clone(in io.Reader, out io.Writer, err io.Writer) RunContext {
	if in == nil {
		in = c.in
	}

	if out == nil {
		out = c.out
	}

	if err == nil {
		err = c.err
	}

	return &context{
		in:    in,
		out:   out,
		err:   err,
		runfn: c.runfn,
	}
}
