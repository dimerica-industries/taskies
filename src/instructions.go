package src

import (
	"fmt"
	"path/filepath"
	"reflect"
)

func decodeInstruction(k string, v reflect.Value) (instruction, error) {
	var ins instruction

	switch k {
	case "set":
		ins = newSetVar()
	case "task":
		ins = newDefineTask()
	case "run":
		ins = newRunTasks()
	case "pipe":
		ins = newRunTasks()
		ins.(*runTasks).pipe = true
	case "include":
		ins = newIncludeNs()
	default:
		ins = newRunTasks()
		v = reflect.ValueOf(map[string]interface{}{k: v.Interface()})
	}

	if err := ins.decode(v); err != nil {
		return nil, err
	}

	return ins, nil
}

type instruction interface {
	decode(reflect.Value) error
	exec(*Runtime, Namespace, *Env) error
}

func newIncludeNs() *includeNs {
	return &includeNs{make([]*_ns, 0)}
}

type includeNs struct {
	ns []*_ns
}

type _ns struct {
	alias string
	path  string
}

func (t *includeNs) decode(data reflect.Value) error {
	k := data.Kind()

	if k == reflect.String {
		t.ns = append(t.ns, &_ns{"", data.String()})
		return nil
	}

	if k == reflect.Map {
		data = reflect.ValueOf([]interface{}{data.Interface()})
		k = reflect.Slice
	}

	if k != reflect.Slice {
		return fmt.Errorf("include directive must be a list")
	}

	l := data.Len()

	for i := 0; i < l; i++ {
		v := data.Index(i).Elem()

		if v.Kind() == reflect.String {
			t.ns = append(t.ns, &_ns{"", v.String()})
			continue
		}

		if v.Kind() != reflect.Map {
			return fmt.Errorf("include item must be a string or map '{alias: path}'")
		}

		keys := v.MapKeys()

		for _, k := range keys {
			vv := v.MapIndex(k).Elem()
			ks := k.String()

			t.ns = append(t.ns, &_ns{ks, vv.String()})
		}
	}

	return nil
}

func (t *includeNs) exec(r *Runtime, ns Namespace, e *Env) error {
	for _, ns1 := range t.ns {
		p, err := filepath.Abs(ns1.path)

		if err != nil {
			return err
		}

		ns2, ast, loaded, err := r.nsg.load(p)

		if err != nil {
			return err
		}

		Debugf("[NS LOAD] [from=%s] [id=%s] [alias=%s] [loaded=%v]", ns.Id(), ns2.Id(), ns1.alias, loaded)

		if !loaded {
			err := inDir(filepath.Dir(p), func() error {
				return execAst(r, ns2, ns2.RootEnv(), ast)
			})

			if err != nil {
				return err
			}
		}

		e.SetVar(ns1.alias, ns2.RootEnv())
	}

	return nil
}

func newDefineTask() *defineTask {
	return &defineTask{
		runList: newRunTasks(),
		set:     newSetVar(),
	}
}

type defineTask struct {
	name        string
	description string
	runList     *runTasks
	set         *setVar
}

func (t *defineTask) decode(data reflect.Value) error {
	if data.Kind() != reflect.Map {
		return invalidTaskType
	}

	keys := data.MapKeys()

	for _, k := range keys {
		ks := k.String()
		v := data.MapIndex(k).Elem()

		switch ks {
		case "name":
			t.name = v.String()
		case "description":
			t.description = v.String()
		case "run":
			t.runList.pipe = false
			if err := t.runList.decode(v); err != nil {
				return err
			}
		case "pipe":
			t.runList.pipe = true
			if err := t.runList.decode(v); err != nil {
				return err
			}
		case "set":
			if err := t.set.decode(v); err != nil {
				return err
			}
		default:
			err := t.runList.decode(reflect.ValueOf(map[string]interface{}{
				ks: v.Interface(),
			}))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *defineTask) exec(r *Runtime, ns Namespace, e *Env) error {
	tsk, err := task(ns, e, t.name, t.description, t.set.vars, t.runList)

	if err != nil {
		return err
	}

	e.AddTask(tsk)
	Debugf("[DEFINE TASK] [NAME=%s] [ENV=%s] %#v", tsk.Name(), e.Id(), tsk)

	return nil
}

func newRunTasks() *runTasks {
	return &runTasks{
		tasks: make([]*runTask, 0),
	}
}

type runTasks struct {
	tasks []*runTask
	pipe  bool
}

func (t *runTasks) decode(data reflect.Value) error {
	if data.Kind() != reflect.Slice {
		t2 := newRunTask()

		if err := t2.decode(data); err != nil {
			return err
		}

		t.tasks = append(t.tasks, t2)
		return nil
	}

	l := data.Len()

	for i := 0; i < l; i++ {
		v := data.Index(i).Elem()
		t2 := newRunTask()

		if err := t2.decode(v); err != nil {
			return err
		}

		t.tasks = append(t.tasks, t2)
	}

	return nil
}

func (t *runTasks) exec(r *Runtime, ns Namespace, e *Env) error {
	tsk, err := task(ns, e, "anon", "", nil, t)

	if err != nil {
		return err
	}

	return r.runWithDefaults(tsk)
}

func newRunTask() *runTask {
	return &runTask{}
}

type runTask struct {
	task    string
	varName string
	args    reflect.Value
}

func (t *runTask) decode(data reflect.Value) error {
	if data.Kind() == reflect.String {
		t.task = data.String()
		return nil
	}

	if data.Kind() != reflect.Map {
		return invalidRunType
	}

	keys := data.MapKeys()

	for _, k := range keys {
		ks := k.String()
		v := data.MapIndex(k).Elem()

		switch ks {
		case "task":
			t.task = v.String()
		case "var":
			t.varName = v.String()
		case "args":
			t.args = v
		default:
			if t.task != "" {
				return invalidRunKey
			}

			t.task = ks
			t.args = v
		}
	}

	return nil
}

func newSetVar() *setVar {
	return &setVar{
		vars: make(map[string]interface{}),
	}
}

type setVar struct {
	vars map[string]interface{}
}

func (t *setVar) decode(data reflect.Value) error {
	if data.Kind() != reflect.Map {
		return invalidSetType
	}

	keys := data.MapKeys()

	for _, k := range keys {
		t.vars[k.String()] = data.MapIndex(k).Elem().Interface()
	}

	return nil
}

func (t *setVar) exec(r *Runtime, ns Namespace, e *Env) error {
	for k, v := range t.vars {
		e.SetVar(k, v)
	}

	return nil
}

func task(ns Namespace, env *Env, name string, description string, export map[string]interface{}, tsks *runTasks) (Task, error) {
	composite := len(tsks.tasks) != 1
	tasks := make([]Task, 0)

	for _, rt := range tsks.tasks {
		name := name
		desc := description

		if composite {
			name = ""
			desc = ""
		}

		switch rt.task {
		case "shell":
			exp := []map[string]interface{}{export}

			if composite {
				exp = make([]map[string]interface{}, 0)
			}

			task := &shellTask{
				baseTask: &baseTask{
					name:        name,
					description: desc,
					export:      exp,
					varName:     rt.varName,
					typ:         "shell",
					env:         env,
				},
				cmd: "sh",
				args: []string{
					"-c",
					rt.args.String(),
				},
			}

			tasks = append(tasks, task)
		default:
			task, _ := ns.RootEnv().GetTask(rt.task)

			if task == nil {
				return nil, fmt.Errorf("Missing task \"%s\"", rt.task)
			}

			exp := task.Export()

			if !composite {
				exp = append(exp, export)
			}

			proxy := proxyTask(task, rt.args)

			proxy.name = name
			proxy.description = desc
			proxy.typ = rt.task
			proxy.varName = rt.varName
			proxy.export = exp
			proxy.env = env

			tasks = append(tasks, proxy)
		}
	}

	var task Task

	if composite {
		if tsks.pipe {
			task = &pipeTask{
				baseTask: &baseTask{
					name:        name,
					description: description,
					typ:         name,
					export:      []map[string]interface{}{export},
					env:         env,
				},
				tasks: tasks,
			}
		} else {
			task = &compositeTask{
				baseTask: &baseTask{
					name:        name,
					description: description,
					typ:         name,
					export:      []map[string]interface{}{export},
					env:         env,
				},
				tasks: tasks,
			}
		}
	} else {
		task = tasks[0]
	}

	return task, nil
}
