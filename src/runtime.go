package src

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

var (
	NsExists   = errors.New("NS already exists")
	RootExists = errors.New("Root ns already exists")
)

func NewRuntime(in io.Reader, out, err io.Writer) *Runtime {
	return &Runtime{
		ns:  make(map[string]*Namespace),
		in:  in,
		out: out,
		err: err,
	}
}

type Runtime struct {
	l    sync.Mutex
	ns   map[string]*Namespace
	root *Namespace
	in   io.Reader
	out  io.Writer
	err  io.Writer
}

func (r *Runtime) Run(task ...string) error {
	ctxt := newContext(r.RootNs().env, r.in, r.out, r.err)

	for _, name := range task {
		t := r.RootNs().ExportedTask(name)

		if t == nil {
			return fmt.Errorf("Task %s does not exist", name)
		}

		res := ctxt.Run(t)

		if res.error != nil {
			return res.error
		}
	}

	return nil
}

func (r *Runtime) RunAll() error {
	return r.Run(r.RootNs().ExportedTasks()...)
}

func (r *Runtime) LoadNs(path string) (*Namespace, error) {
	r.l.Lock()
	defer r.l.Unlock()

	id, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	if _, ok := r.ns[id]; ok {
		return nil, fmt.Errorf("NS with id already exists")
	}

	contents, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	ns, err := r.addNs(id)

	if err != nil {
		return nil, err
	}

	err = DecodeYAML(contents, ns)

	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (r *Runtime) RootNs() *Namespace {
	r.l.Lock()
	defer r.l.Unlock()

	return r.root
}

func (r *Runtime) addNs(id string) (*Namespace, error) {
	if _, ok := r.ns[id]; ok {
		return nil, NsExists
	}

	ns := &Namespace{
		id:              id,
		exportedTasks:   make(map[string]Task),
		unexportedTasks: make(map[string]Task),
		env:             NewEnv(),
	}

	r.ns[id] = ns

	if r.root == nil {
		r.root = ns
	}

	return ns, nil
}

func (r *Runtime) AddNs(id string) (*Namespace, error) {
	r.l.Lock()
	defer r.l.Unlock()

	return r.addNs(id)
}

func (r *Runtime) Ns(id string) *Namespace {
	r.l.Lock()
	defer r.l.Unlock()

	return r.ns[id]
}

type Namespace struct {
	l               sync.Mutex
	id              string
	exportedTasks   map[string]Task
	unexportedTasks map[string]Task
	env             *Env
}

func (ns *Namespace) ExportedTasks() []string {
	m := make([]string, len(ns.exportedTasks))
	i := 0

	for k, _ := range ns.exportedTasks {
		m[i] = k
		i++
	}

	return m
}

func (ns *Namespace) AddTask(t Task) error {
	ns.l.Lock()
	defer ns.l.Unlock()

	n := t.Name()
	ln := strings.ToLower(n)
	m := ns.exportedTasks

	if ln[0] == n[0] {
		m = ns.unexportedTasks
	}

	if _, ok := m[ln]; ok {
		return fmt.Errorf("Task with name %s already exists", n)
	}

	m[ln] = t

	return nil
}

func (ns *Namespace) ExportedTask(name string) Task {
	ns.l.Lock()
	defer ns.l.Unlock()

	ln := strings.ToLower(name)
	return ns.exportedTasks[ln]
}

func (ns *Namespace) Task(name string) Task {
	ns.l.Lock()
	defer ns.l.Unlock()

	ln := strings.ToLower(name)
	m := ns.exportedTasks

	if ln[0] == name[0] {
		m = ns.unexportedTasks
	}

	return m[ln]
}
