package src

import (
	"io/ioutil"
	"path/filepath"
	"sync"
)

func newLoader() *loader {
	return &loader{
		loads: make(map[string]*load),
	}
}

type loader struct {
	l     sync.Mutex
	loads map[string]*load
}

type load struct {
	id  string
	raw []byte
	ast *ast
}

func (l *loader) load(path string) (*load, bool, error) {
	l.l.Lock()
	defer l.l.Unlock()

	id, err := filepath.Abs(path)

	if err != nil {
		return nil, false, err
	}

	if load, ok := l.loads[id]; ok {
		return load, true, nil
	}

	raw, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, false, err
	}

	ast, err := parseBytes(raw)

	if err != nil {
		return nil, false, err
	}

	l.loads[id] = &load{
		id:  id,
		raw: raw,
		ast: ast,
	}

	return l.loads[id], false, nil
}
