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

func (t *baseTask) set(name, description, typ, varName string, export []map[string]interface{}, env *Env) {

}

type funcTask struct {
	*baseTask
	fn func(r RunContext) error
}

func (t *funcTask) Run(r RunContext) error {
	return t.fn(r)
}

func proxyTask(task Task, args reflect.Value) *funcTask {
	return &funcTask{
		baseTask: &baseTask{},
		fn: func(r RunContext) error {
			env := r.Env()
			env.addParent(task.Env())

			if args.Kind() == reflect.Map {
				keys := args.MapKeys()

				for _, k := range keys {
					env.SetVar(k.String(), args.MapIndex(k).Elem().Interface())
				}
			}

			r2 := r.Clone(nil, nil, nil, env)

			return task.Run(r2)
		},
	}
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

	Debugf("[SHELL] [ENV=%s] %s %s", r.Env().Id(), cmd, args)
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
