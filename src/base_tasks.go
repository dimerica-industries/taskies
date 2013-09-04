package src

import (
	"io"
	"os/exec"
	"reflect"
)

type proxyTask struct {
	name        string
	description string
	typ         string
	varName     string
	task        Task
	export      []map[string]interface{}
	args        reflect.Value
}

func (t *proxyTask) Name() string {
	return t.name
}

func (t *proxyTask) Description() string {
	return t.description
}

func (t *proxyTask) Export() []map[string]interface{} {
	return t.export
}

func (t *proxyTask) Type() string {
	return t.typ
}

func (t *proxyTask) Var() string {
	return t.varName
}

func (t *proxyTask) Run(r RunContext) error {
	if t.args.Kind() == reflect.Map {
		keys := t.args.MapKeys()

		for _, k := range keys {
			r.Env().SetVar(k.String(), t.args.MapIndex(k).Elem().Interface())
		}
	}

	return t.task.Run(r)
}

type shellTask struct {
	name        string
	varName     string
	description string
	cmd         string
	args        []string
	export      []map[string]interface{}
}

func (t *shellTask) Type() string {
	return "shell"
}

func (t *shellTask) Var() string {
	return t.varName
}

func (t *shellTask) Name() string {
	return t.name
}

func (t *shellTask) Description() string {
	return t.description
}

func (t *shellTask) Export() []map[string]interface{} {
	return t.export
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
	name        string
	varName     string
	description string
	tasks       []Task
	typ         string
	export      []map[string]interface{}
}

func (t *compositeTask) Type() string {
	return t.typ
}

func (t *compositeTask) Var() string {
	return t.varName
}

func (t *compositeTask) Name() string {
	return t.name
}

func (t *compositeTask) Description() string {
	return t.description
}

func (t *compositeTask) Export() []map[string]interface{} {
	return t.export
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
	name        string
	varName     string
	description string
	typ         string
	export      []map[string]interface{}
	tasks       []Task
}

func (t *pipeTask) Var() string {
	return t.varName
}

func (t *pipeTask) Name() string {
	return t.name
}

func (t *pipeTask) Description() string {
	return t.description
}

func (t *pipeTask) Export() []map[string]interface{} {
	return t.export
}

func (t *pipeTask) Type() string {
	return t.typ
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
			r = r.Clone(in, out, nil)
			err := r.Run(t)

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
