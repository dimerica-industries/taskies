package src

import (
	"strings"
	"sync"
)

func NewEnv() *Env {
	return &Env{
		vars:          newVarSet(),
		tasks:         make([]string, 0),
		exportedTasks: make([]string, 0),
        exportedTasksMap: make(map[string]bool),
	}
}

type Env struct {
	parent        *Env
	vars          *varSet
	taskLock      sync.Mutex
	tasks         []string
	exportedTasks []string
    exportedTasksMap map[string]bool
}

func (e *Env) GetVar(k string) interface{} {
	v := e.vars.get(k)

	if v != nil || e.IsRoot() {
		return v
	}

	return e.parent.GetVar(k)
}

func (e *Env) SetVar(k string, v interface{}) {
	k = template(k, e).(string)

	if ev, ok := v.(*Env); ok {
		Debugf("[ENV SET VAR] [ENV=%p] [KEY=%#v] [VALUE=%p]", e, k, v)
		v = ev.vars
	} else {
		Debugf("[ENV SET VAR] [ENV=%p] [KEY=%#v] [VALUE=%#v]", e, k, v)
		v = template(v, e)
	}

	e.vars.set(k, v)
}

func (e *Env) Tasks() []string {
	return e.tasks
}

func (e *Env) ExportedTasks() []string {
	return e.exportedTasks
}

func (e *Env) GetTask(name string) Task {
	t := e.vars.Get(name)
	root := e.IsRoot()

	if t == nil {
		if root {
			return nil
		}

		return e.parent.GetTask(name)
	}

	tsk := t.(Task)

	if tsk != nil || e.IsRoot() {
		return tsk
	}

	return e.parent.GetTask(name)
}

func (e *Env) GetExportedTask(name string) Task {
    e.taskLock.Lock()
    defer e.taskLock.Unlock()

    if _, ok := e.exportedTasksMap[name]; ok {
        return e.GetTask(name)
    }

    return nil
}

func (e *Env) AddTask(t Task) {
	e.taskLock.Lock()
	defer e.taskLock.Unlock()

	name := t.Name()
	lname := strings.ToLower(name)

	if name[0] != lname[0] {
		e.exportedTasks = append(e.exportedTasks, name)
        e.exportedTasksMap[name] = true
	}

	e.tasks = append(e.tasks, name)

	e.vars.Set(name, t)
}

func (e *Env) Child() *Env {
	e2 := NewEnv()
	e2.parent = e

	Debugf("[ENV CHILD] [parent=%p] [child=%p]", e, e2)

	return e2
}

func (e *Env) IsRoot() bool {
	return e.parent == nil
}

func (e *Env) Root() *Env {
	if e.IsRoot() {
		return e
	}

	return e.parent.Root()
}
