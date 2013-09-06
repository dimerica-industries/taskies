package src

import (
	"io"
	"os/exec"
	"reflect"
)

type baseTask struct {
	name        string
	description string
	typ         string
	varName     string
	export      []map[string]interface{}
	env         *Env
}

func (t *baseTask) Name() string {
	return t.name
}

func (t *baseTask) Description() string {
	return t.description
}

func (t *baseTask) Export() []map[string]interface{} {
	return t.export
}

func (t *baseTask) Type() string {
	return t.typ
}

func (t *baseTask) Var() string {
	return t.varName
}

func (t *baseTask) Env() *Env {
	return t.env
}

type proxyTask struct {
	*baseTask
	task Task
	args reflect.Value
}

func (t *proxyTask) Run(r RunContext) error {
	if t.args.Kind() == reflect.Map {
		keys := t.args.MapKeys()

		for _, k := range keys {
			r.Env().SetVar(k.String(), t.args.MapIndex(k).Elem().Interface())
		}
	}

    env := r.Env().Child()
    env.addParent(t.task.Env())

	r2 := r.Clone(nil, nil, nil, env)

	return t.task.Run(r2)
}

type shellTask struct {
	*baseTask
	cmd  string
	args []string
}

func (t *shellTask) Run(r RunContext) error {
	cmd := template(t.cmd, r.Env()).(string)
	args := make([]string, len(t.args))

	for i, _ := range args {
		args[i] = template(t.args[i], r.Env()).(string)
	}

	Debugf("[SHELL] %s %s", cmd, args)
	c := exec.Command(cmd, args...)

	c.Stdin = r.In()
	c.Stdout = r.Out()
	c.Stderr = r.Err()

	return c.Run()
}

type compositeTask struct {
	*baseTask
	tasks []Task
}

func (t *compositeTask) Run(r RunContext) error {
	for _, tt := range t.tasks {
		if err := r.Run(tt); err != nil {
			return err
		}
	}

	return nil
}

type pipeTask struct {
	*baseTask
	tasks []Task
}

func (t *pipeTask) Run(r RunContext) error {
	ch := make(chan error)
	l := len(t.tasks)
	in := r.In()

	for i, t := range t.tasks {
		var (
			pin io.Reader
			out io.Writer
		)

		if i == l-1 {
			out = r.Out()
		} else {
			pin, out = io.Pipe()
		}

		go func(t Task, in io.Reader, out io.Writer) {
			r2 := r.Clone(in, out, nil, nil)
			err := r2.Run(t)

			if c, ok := out.(io.Closer); ok {
				c.Close()
			}

			ch <- err
		}(t, in, out)

		in = pin
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
