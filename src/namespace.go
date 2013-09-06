package src

import (
	"sync"
)

type Namespace interface {
	Id() string
	RootEnv() *Env
	Tasks() []string
	GetTask(string) Task
}

func newNs(id string) *ns {
	return &ns{
		id:  id,
		env: NewEnv(),
	}
}

type ns struct {
	id  string
	env *Env
}

func (n *ns) Id() string {
	return n.id
}

func (n *ns) Tasks() []string {
	return n.RootEnv().ExportedTasks()
}

func (n *ns) GetTask(k string) Task {
	t, _ := n.RootEnv().GetExportedTask(k)
	return t
}

func (n *ns) RootEnv() *Env {
	return n.env
}

func newNsGroup(l *loader) *nsGroup {
	return &nsGroup{
		loader: l,
		ns:     make(map[string]Namespace),
	}
}

type nsGroup struct {
	sync.Mutex
	loader *loader
	ns     map[string]Namespace
}

func (n *nsGroup) load(path string) (Namespace, *ast, error) {
	Debugf("[LOADING] %s", path)

	n.Lock()
	defer n.Unlock()

	l, err := n.loader.load(path)

	if err != nil {
		return nil, nil, err
	}

	if ns, ok := n.ns[l.id]; ok {
		return ns, nil, nil
	}

	ns := newNs(l.id)

	return ns, l.ast, nil
}

func (n *nsGroup) get(k string) Namespace {
	n.Lock()
	defer n.Unlock()

	return n.ns[k]
}

func (n *nsGroup) add(ns Namespace) {
	n.Lock()
	defer n.Unlock()

	if _, ok := n.ns[ns.Id()]; ok {
		return
	}

	n.ns[ns.Id()] = ns
}
