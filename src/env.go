package src

import (
	"fmt"
	"strings"
	"sync"
)

func NewEnv() *Env {
	return &Env{
		vars:             newVarSet(),
		tasks:            make([]string, 0),
		exportedTasks:    make([]string, 0),
		exportedTasksMap: make(map[string]bool),
	}
}

type Env struct {
	parents          []*Env
	vars             *varSet
	taskLock         sync.Mutex
	tasks            []string
	exportedTasks    []string
	exportedTasksMap map[string]bool
}

func (e *Env) Id() string {
	str := fmt.Sprintf("%p", e)

	if len(e.parents) > 0 {
		parts := make([]string, len(e.parents))

		for i, p := range e.parents {
			parts[i] = p.Id()
		}

		str += " < (" + strings.Join(parts, ", ") + ")"
	}

	return str
}

func (e *Env) GetVar(k string) interface{} {
	v := e.vars.get(k)

	if v != nil || e.IsRoot() {
		return v
	}

	for _, p := range e.parents {
		if v = p.GetVar(k); v != nil {
			return v
		}
	}

	return nil
}

func (e *Env) SetVar(k string, v interface{}) {
	k = template(k, e).(string)

	if ev, ok := v.(*Env); ok {
		Debugf("[ENV SET VAR] [ENV=%s] [KEY=%#v] [VALUE=%s]", e.Id(), k, ev.Id())
		v = ev.vars
	} else {
		Debugf("[ENV SET VAR] [ENV=%s] [KEY=%#v] [VALUE=%#v]", e.Id(), k, v)
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

func (e *Env) GetTask(name string) (Task, *Env) {
	t := e.vars.Get(name)
	root := e.IsRoot()

	tryParents := func() (Task, *Env) {
		for _, p := range e.parents {
			if t, e := p.GetTask(name); t != nil {
				return t, e
			}
		}

		return nil, nil
	}

	if t == nil {
		if root {
			return nil, nil
		}

		return tryParents()
	}

	tsk := t.(Task)

	if tsk != nil || e.IsRoot() {
		return tsk, e
	}

	return tryParents()
}

func (e *Env) GetExportedTask(name string) (Task, *Env) {
	e.taskLock.Lock()
	defer e.taskLock.Unlock()

	if _, ok := e.exportedTasksMap[name]; ok {
		return e.GetTask(name)
	}

	return nil, nil
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
	e2.addParent(e)

	return e2
}

func (e *Env) IsRoot() bool {
	return len(e.parents) == 0
}

func (e *Env) addParent(p *Env) {
	e.parents = append(e.parents, p)
	Debugf("[ENV PARENT] [parent=%p] [child=%s]", p, e.Id())
}
