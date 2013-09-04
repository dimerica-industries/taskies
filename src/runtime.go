package src

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	MissingTask = errors.New("Task doesn't exist")
	TaskExists  = errors.New("Task with same name exists")
)

func LoadRuntime(path string, in io.Reader, out, err io.Writer) (*Runtime, error) {
	rt := newRuntime(in, out, err)

	ns, ast, e := rt.nsg.load(path)

	if e != nil {
		return nil, e
	}

	p := filepath.Dir(path)
	Debugf("[CHDIR] %s", p)

	if e := os.Chdir(p); e != nil {
		return nil, e
	}

	rt.ns = ns

	if err := execAst(rt, ns, ns.RootEnv(), ast); err != nil {
		return nil, err
	}

	return rt, nil
}

func NewRuntime(in io.Reader, out, err io.Writer) *Runtime {
	rt := newRuntime(in, out, err)

	ns := newNs("__root__")
	rt.nsg.add(ns)
	rt.ns = ns

	return rt
}

func newRuntime(in io.Reader, out, err io.Writer) *Runtime {
	loader := newLoader()
	nsg := newNsGroup(loader)

	return &Runtime{
		nsg: nsg,
		in:  in,
		out: out,
		err: err,
	}
}

type Runtime struct {
	ns  Namespace
	nsg *nsGroup
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (r *Runtime) In() io.Reader {
	return r.in
}

func (r *Runtime) Out() io.Writer {
	return r.out
}

func (r *Runtime) Err() io.Writer {
	return r.err
}

func (r *Runtime) Run(task string) error {
	t := r.ns.RootEnv().GetTask(task)

	if t == nil {
		return MissingTask
	}

	return r.runWithDefaults(t)
}

func (r *Runtime) RootNs() Namespace {
	return r.ns
}

func (r *Runtime) runWithDefaults(t Task) error {
	return r.run(t, r.ns.RootEnv(), r.In(), r.Out(), r.Err())
}

func (r *Runtime) run(t Task, env *Env, in io.Reader, out, err io.Writer) error {
	name := t.Name()

	if name == "" {
		name = t.Type()
	}

	if t2 := env.GetVar("TASKS." + name); t2 != nil {
		i := 1

		for {
			n := fmt.Sprintf("%s_%d", name, i)

			if t2 := env.GetVar("TASKS." + n); t2 == nil {
				name = n
				break
			}

			i++
		}
	}

	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)

	sout := io.MultiWriter(out, bout)
	serr := io.MultiWriter(err, berr)

	cenv := env.Child()

	ctxt := &context{
		in:  in,
		out: sout,
		err: serr,
		runfn: func(c RunContext, t2 Task) error {
			return r.run(t2, cenv, c.In(), c.Out(), c.Err())
		},
		env: cenv,
	}

	e := t.Run(ctxt)

	if e != nil {
		cenv.SetVar("ERROR", e.Error())
	}

	cenv.SetVar("OUT", strings.TrimRightFunc(string(bout.Bytes()), unicode.IsSpace))
	cenv.SetVar("ERR", strings.TrimRightFunc(string(berr.Bytes()), unicode.IsSpace))

	exp := t.Export()

	for _, vars := range exp {
		for k, v := range vars {
			cenv.SetVar(k, v)
		}
	}

	env.SetVar("LAST", cenv)
	env.SetVar("TASKS."+name, cenv)

	if t.Var() != "" {
		env.SetVar(t.Var(), cenv)
	}

	return e
}

func execAst(r *Runtime, ns Namespace, e *Env, a *ast) error {
	for _, ins := range a.instructions {
		if err := ins.exec(r, ns, e); err != nil {
			return err
		}
	}

	return nil
}
